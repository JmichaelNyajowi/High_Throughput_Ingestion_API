package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/example/telemetry/api/internal/platform/httpserver"
	"github.com/example/telemetry/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
	redis      *redis.Client
	publisher  *telemetry.RedisStatePublisher
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
	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		pool.Close()
		return nil, err
	}
	redisClient := redis.NewClient(redisOptions)
	redisTimeout := cfg.Dependencies.RedisTimeout
	if redisTimeout > 250*time.Millisecond {
		redisTimeout = 250 * time.Millisecond
	}
	publisher, err := telemetry.NewRedisStatePublisher(redisClient, telemetry.RedisPublisherConfig{Timeout: redisTimeout, Logger: logger})
	if err != nil {
		_ = redisClient.Close()
		pool.Close()
		return nil, err
	}

	repository := telemetry.NewPostgresEventRepository(pool)
	retry, err := telemetry.NewRetryBuffer(repository, telemetry.RetryConfig{})
	if err != nil {
		_ = publisher.Shutdown(context.Background())
		_ = redisClient.Close()
		pool.Close()
		return nil, err
	}
	batcher, err := telemetry.NewBatcher(repository, persistenceInputCapacityBatches, retry)
	if err != nil {
		_ = retry.Shutdown(context.Background())
		_ = publisher.Shutdown(context.Background())
		_ = redisClient.Close()
		pool.Close()
		return nil, err
	}
	aggregates := telemetry.NewAggregateEngine(nil)
	queue, err := telemetry.NewDeviceShardedQueue(telemetry.QueueConfig{
		ShardCount:       cfg.AdmissionQueue.ShardCount,
		CapacityPerShard: cfg.AdmissionQueue.CapacityPerShard,
	}, func(ctx context.Context, batch ingestion.ValidatedBatch) {
		snapshots := aggregates.Process(batch)
		publisher.Publish(batch.Device.ExternalID, time.Now().UTC(), snapshots)
		batcher.Processor(ctx, batch)
	}, logger)
	if err != nil {
		_ = batcher.Shutdown(context.Background())
		_ = retry.Shutdown(context.Background())
		_ = publisher.Shutdown(context.Background())
		_ = redisClient.Close()
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
		redis:      redisClient,
		publisher:  publisher,
	}, nil
}

func (application *application) Shutdown(ctx context.Context) error {
	var firstError error
	for _, drainer := range []interface{ Shutdown(context.Context) error }{application.queue, application.publisher, application.batcher, application.retry} {
		if err := drainer.Shutdown(ctx); err != nil && firstError == nil {
			firstError = err
		}
	}
	if err := application.redis.Close(); err != nil && firstError == nil {
		firstError = err
	}
	application.pool.Close()
	return firstError
}

func (application *application) pauseAdmission() {
	application.readiness.Pause()
	application.queue.Pause()
}
