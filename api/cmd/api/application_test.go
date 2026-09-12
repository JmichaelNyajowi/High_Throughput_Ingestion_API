package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17.3-alpine", postgres.WithDatabase("telemetry"), postgres.WithUsername("telemetry"), postgres.WithPassword("test-password"), postgres.WithInitScripts(filepath.Join("..", "..", "..", "migrations", "000001_init.up.sql")))
	if err != nil {
		t.Fatalf("start PostgreSQL: %v", err)
	}
	testcontainers.CleanupContainer(t, container)
	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	waitForApplicationPostgres(t, ctx, pool)

	const pepper = "12345678901234567890123456789012"
	apiKey := applicationTestAPIKey("edge-01", 1)
	if _, err := credentials.Bootstrap(ctx, pool, pepper, []config.DeviceSeed{{DeviceID: "edge-01", APIKey: apiKey}}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	application, err := newApplication(ctx, config.Config{
		PostgresURL:  url,
		APIKeyPepper: pepper,
		AdmissionQueue: config.AdmissionQueueConfig{
			ShardCount:       1,
			CapacityPerShard: 1,
		},
	}, nil)
	if err != nil {
		t.Fatalf("create application: %v", err)
	}
	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownContext); err != nil {
			t.Errorf("shutdown application: %v", err)
		}
	})

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := application.retry.Shutdown(shutdownContext); err != nil {
		t.Fatalf("disable retry guard: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/telemetry/batches", strings.NewReader(`{"events":[{"event_id":"event-01","device_id":"edge-01","timestamp":"2026-09-12T00:00:00Z","measurement_type":"temperature","value":20}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", apiKey)
	response := httptest.NewRecorder()

	application.handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("response = %d, headers=%#v; want 503 with request ID", response.Code, response.Header())
	}
	if depth := application.queue.Metrics().Depth; depth != 0 {
		t.Fatalf("queue depth = %d, want no enqueue under retry backpressure", depth)
	}
}

func waitForApplicationPostgres(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := pool.Ping(ctx); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("PostgreSQL did not become ready within 30 seconds")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func applicationTestAPIKey(keyID string, value byte) string {
	secret := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string([]byte{value}), 32)))
	return "tk_" + keyID + "_" + secret
}
