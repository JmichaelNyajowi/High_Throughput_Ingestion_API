package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/example/telemetry/api/internal/fleet"
	"github.com/example/telemetry/api/internal/history"
	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/example/telemetry/api/internal/platform/httpserver"
	"github.com/example/telemetry/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	fleet      *fleet.Reader
	rebuilder  *telemetry.RedisRebuilder
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
	rebuilder, err := telemetry.NewRedisRebuilder(telemetry.NewPostgresRebuildSource(pool), publisher)
	if err != nil {
		_ = publisher.Shutdown(context.Background())
		_ = redisClient.Close()
		pool.Close()
		return nil, err
	}
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
	fleetReader := fleet.New(redisClient, func() string { return string(publisher.Mode()) })
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_queue_capacity", Help: "Configured admission queue capacity."}, func() float64 { return float64(queue.Metrics().Capacity) }))
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_queue_depth", Help: "Current admission queue depth."}, func() float64 { return float64(queue.Metrics().Depth) }))
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_workers", Help: "Fixed queue worker count."}, func() float64 { return float64(queue.Metrics().WorkerCount) }))
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_redis_errors_total", Help: "Redis publication errors."}, func() float64 { return float64(publisher.Metrics().Errors) }))
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_persistence_flushes_total", Help: "Persistence flushes."}, func() float64 { return float64(batcher.Metrics().Flushes) }))
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "telemetry_retry_occupancy_events", Help: "Retry-buffer event occupancy."}, func() float64 { return float64(retry.Metrics().OccupancyEvents) }))
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
			Metrics:   promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
			ReadinessCheck: func() (string, string, string, string) {
				p := "ready"
				c, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				if pool.Ping(c) != nil {
					p = "unavailable"
				}
				a := "accepting"
				s := "ready"
				if !retry.AllowsAdmission() {
					a = "backpressured"
					s = "unavailable"
				}
				if p != "ready" {
					s = "unavailable"
				}
				return a, p, string(publisher.Mode()), s
			},
			RegisterRoutes: func(mux *http.ServeMux) {
				mux.Handle("POST /v1/telemetry/batches", admissionRoute)
				mux.HandleFunc("GET /v1/live/fleet", func(w http.ResponseWriter, r *http.Request) {
					devices, err := fleetReader.Fleet(r.Context())
					if err != nil {
						httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Live aggregates are unavailable")
						return
					}
					active, stale, warning, critical := 0, 0, 0, 0
					for _, d := range devices {
						if d.Freshness == "active" {
							active++
						} else {
							stale++
						}
						if d.OverallStatus == "warning" {
							warning++
						}
						if d.OverallStatus == "critical" {
							critical++
						}
					}
					httpserver.WriteJSON(w, http.StatusOK, map[string]any{"mode": "live", "generated_at": time.Now().UTC(), "summary": map[string]int{"active": active, "stale": stale, "warning": warning, "critical": critical}, "devices": devices})
				})
				mux.HandleFunc("GET /v1/live/devices/{deviceId}", func(w http.ResponseWriter, r *http.Request) {
					id := r.PathValue("deviceId")
					d, err := fleetReader.Device(r.Context(), id)
					if errors.Is(err, fleet.ErrDeviceNotFound) {
						httpserver.WriteError(w, r, http.StatusNotFound, "not_found", "Device has no current live aggregate")
						return
					}
					if err != nil {
						httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Live aggregates are unavailable")
						return
					}
					httpserver.WriteJSON(w, http.StatusOK, map[string]any{"mode": "live", "generated_at": time.Now().UTC(), "device": d})
				})
				mux.HandleFunc("GET /v1/history/devices/{deviceId}", func(w http.ResponseWriter, r *http.Request) {
					from, e1 := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
					to, e2 := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
					limit, e3 := strconv.Atoi(r.URL.Query().Get("limit"))
					if e1 != nil || e2 != nil || e3 != nil {
						httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid history bounds")
						return
					}
					measurement := r.URL.Query().Get("measurement_type")
					events, err := history.Query(r.Context(), pool, r.PathValue("deviceId"), from, to, limit, measurement)
					if err != nil {
						httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid history bounds")
						return
					}
					httpserver.WriteJSON(w, http.StatusOK, map[string]any{"request_id": httpserver.RequestID(r.Context()), "events": events, "limit": limit})
				})
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
		fleet:      fleetReader,
		rebuilder:  rebuilder,
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
