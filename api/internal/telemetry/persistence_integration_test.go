package telemetry

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresEventRepositoryFlushesTransactionallyAndIgnoresDuplicateEvents(t *testing.T) {
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
	if err := pool.Ping(ctx); err != nil {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(250 * time.Millisecond)
			if err = pool.Ping(ctx); err == nil {
				break
			}
		}
		if err != nil {
			t.Fatalf("wait for PostgreSQL: %v", err)
		}
	}
	var deviceID string
	if err := pool.QueryRow(ctx, `INSERT INTO devices (external_id) VALUES ('edge-01') RETURNING id`).Scan(&deviceID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	batch := ingestion.ValidatedBatch{Device: credentials.AuthorizedDevice{InternalID: deviceID, ExternalID: "edge-01"}, RequestID: "00000000-0000-4000-8000-000000000001", Events: []ingestion.ValidatedEvent{{EventID: "event-01", MeasurementType: "temperature", Value: 20, Unit: "C", Timestamp: now}}}
	outcome, err := NewPostgresEventRepository(pool).Flush(ctx, []ingestion.ValidatedBatch{batch, batch})
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if outcome.Attempted != 2 || outcome.Inserted != 1 || outcome.Conflicts != 1 {
		t.Fatalf("outcome=%#v", outcome)
	}
	var rows int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM telemetry_events WHERE device_id = $1 AND event_id = 'event-01'`, deviceID).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("rows=%d", rows)
	}
}

func TestBatcherFlushesToPostgreSQLOnTimerAndSize(t *testing.T) {
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
	if err := pool.QueryRow(ctx, `INSERT INTO devices (external_id) VALUES ('edge-batcher') RETURNING id`).Scan(&deviceID); err != nil {
		t.Fatal(err)
	}
	batcher, err := NewBatcher(NewPostgresEventRepository(pool), 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := batcher.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown batcher: %v", err)
		}
	})

	now := time.Now().UTC().Truncate(time.Microsecond)
	timerBatch := persistenceBatch(deviceID, "timer-event", "00000000-0000-4000-8000-000000000002", now, 1)
	if !batcher.Submit(timerBatch) {
		t.Fatal("submit timer batch")
	}
	waitForEventCount(t, ctx, pool, deviceID, 1)

	sizeBatch := persistenceBatch(deviceID, "size-event", "00000000-0000-4000-8000-000000000003", now, PersistenceFlushSize)
	if !batcher.Submit(sizeBatch) {
		t.Fatal("submit size batch")
	}
	waitForEventCount(t, ctx, pool, deviceID, PersistenceFlushSize+1)

	metrics := batcher.Metrics()
	if metrics.Flushes < 2 || metrics.Failures != 0 || metrics.LastFlush.Attempted != PersistenceFlushSize || metrics.LastFlush.Inserted != PersistenceFlushSize || metrics.LastFlushDuration <= 0 {
		t.Fatalf("metrics=%#v", metrics)
	}
}

func persistenceBatch(deviceID, eventPrefix, requestID string, timestamp time.Time, eventCount int) ingestion.ValidatedBatch {
	events := make([]ingestion.ValidatedEvent, eventCount)
	for i := range events {
		events[i] = ingestion.ValidatedEvent{EventID: fmt.Sprintf("%s-%04d", eventPrefix, i), MeasurementType: "temperature", Value: 20, Unit: "C", Timestamp: timestamp}
	}
	return ingestion.ValidatedBatch{Device: credentials.AuthorizedDevice{InternalID: deviceID}, RequestID: requestID, Events: events}
}

func waitForPostgres(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
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

func waitForEventCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, deviceID string, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM telemetry_events WHERE device_id = $1`, deviceID).Scan(&count); err == nil && count == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("telemetry row count did not become %d", want)
		}
		time.Sleep(25 * time.Millisecond)
	}
}
