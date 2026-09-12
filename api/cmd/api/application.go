package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/example/telemetry/api/internal/platform/httpserver"
	"github.com/example/telemetry/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A full 500-event request must not grow the persistence input beyond the
// retry-buffer safety boundary while a failed flush is being retried.
const persistenceInputCapacityBatches = telemetry.RetryBufferCapacityEvents / 500

type application struct {
	handler    http.Handler
	readiness  *httpserver.Readiness
	pool       *pgxpool.Pool
	queue      *telemetry.DeviceShardedQueue
	batcher    *telemetry.Batcher
	retry      *telemetry.RetryBuffer
	aggregates *telemetry.AggregateEngine
}

func newApplication(ctx context.Context, cfg config.Config, logger *slog.Logger) (*application, error) {
	if logger == nil {
		logger = slog.Default()
	}
	pool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	repository := telemetry.NewPostgresEventRepository(pool)
	retry, err := telemetry.NewRetryBuffer(repository, telemetry.RetryConfig{})
	if err != nil {
		pool.Close()
		return nil, err
	}
	batcher, err := telemetry.NewBatcher(repository, persistenceInputCapacityBatches, retry)
	if err != nil {
		_ = retry.Shutdown(context.Background())
		pool.Close()
		return nil, err
	}
	aggregates := telemetry.NewAggregateEngine(nil)
	queue, err := telemetry.NewDeviceShardedQueue(telemetry.QueueConfig{
		ShardCount:       cfg.AdmissionQueue.ShardCount,
		CapacityPerShard: cfg.AdmissionQueue.CapacityPerShard,
	}, func(ctx context.Context, batch ingestion.ValidatedBatch) {
		aggregates.Process(batch)
		batcher.Processor(ctx, batch)
	}, logger)
	if err != nil {
		_ = batcher.Shutdown(context.Background())
		_ = retry.Shutdown(context.Background())
		pool.Close()
		return nil, err
	}

	readiness := httpserver.NewReadiness()
	admissionRoute := ingestion.NewTelemetryAdmissionRoute(
		ingestion.NewDeviceAuthenticator(pool, cfg.APIKeyPepper),
		ingestion.NewBatchValidator(nil),
		ingestion.NewDeviceRateLimiter(nil),
		queue,
		retry,
	)
	return &application{
		handler: httpserver.NewHandler(httpserver.Options{
			Logger:    logger,
			Readiness: readiness,
			RegisterRoutes: func(mux *http.ServeMux) {
				mux.Handle("POST /v1/telemetry/batches", admissionRoute)
			},
		}),
		readiness:  readiness,
		pool:       pool,
		queue:      queue,
		batcher:    batcher,
		retry:      retry,
		aggregates: aggregates,
	}, nil
}

func (application *application) Shutdown(ctx context.Context) error {
	var firstError error
	for _, drainer := range []interface{ Shutdown(context.Context) error }{application.queue, application.batcher, application.retry} {
		if err := drainer.Shutdown(ctx); err != nil && firstError == nil {
			firstError = err
		}
	}
	application.pool.Close()
	return firstError
}

func (application *application) pauseAdmission() {
	application.readiness.Pause()
	application.queue.Pause()
}
