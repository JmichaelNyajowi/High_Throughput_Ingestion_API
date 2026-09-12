package telemetry

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRetryBufferActivatesBackpressureDuringPostgreSQLOutage(t *testing.T) {
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
	waitForPostgres(t, ctx, pool)

	var deviceID string
	if err := pool.QueryRow(ctx, `INSERT INTO devices (external_id) VALUES ('edge-outage') RETURNING id`).Scan(&deviceID); err != nil {
		t.Fatal(err)
	}
	if err := container.Stop(ctx, nil); err != nil {
		t.Fatalf("stop PostgreSQL: %v", err)
	}

	buffer, err := NewRetryBuffer(NewPostgresEventRepository(pool), RetryConfig{CapacityEvents: RetryBufferCapacityEvents})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := buffer.Shutdown(shutdownContext); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("shutdown retry buffer: %v", err)
		}
	})
	if !buffer.Enqueue([]ingestion.ValidatedBatch{persistenceBatch(deviceID, "outage-event", "00000000-0000-4000-8000-000000000004", time.Now().UTC(), 1)}) {
		t.Fatal("enqueue outage batch")
	}
	waitForRetryCondition(t, func() bool { return buffer.Metrics().Exhausted == 1 })
	if metrics := buffer.Metrics(); metrics.Retries != RetryMaximumAttempts || metrics.OccupancyEvents != 0 || !metrics.Backpressured {
		t.Fatalf("outage retry metrics=%#v", metrics)
	}
	if buffer.AllowsAdmission() {
		t.Fatal("PostgreSQL outage exhaustion must block admission")
	}
}
