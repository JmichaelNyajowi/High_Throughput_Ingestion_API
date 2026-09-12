package telemetry

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxRebuildRange  = 15 * time.Minute
	maxRebuildEvents = 10_000
)

var ErrRebuildBusy = errors.New("redis rebuild is already running")

type RebuildEvent struct {
	DeviceID   string
	Event      ingestion.ValidatedEvent
	ReceivedAt time.Time
}

// RebuildSource supplies an explicitly bounded, received-time-ordered scan.
type RebuildSource interface {
	LoadForRebuild(context.Context, time.Time, time.Time, int) ([]RebuildEvent, error)
}

type PostgresRebuildSource struct{ pool *pgxpool.Pool }

func NewPostgresRebuildSource(pool *pgxpool.Pool) PostgresRebuildSource {
	return PostgresRebuildSource{pool: pool}
}

func (source PostgresRebuildSource) LoadForRebuild(ctx context.Context, from, to time.Time, limit int) ([]RebuildEvent, error) {
	if source.pool == nil || limit < 1 || limit > maxRebuildEvents {
		return nil, errors.New("bounded postgres rebuild source is required")
	}
	rows, err := source.pool.Query(ctx, `SELECT d.external_id, e.event_id, e.measurement_type, e.value, e.unit, e.event_timestamp, e.received_at FROM telemetry_events e JOIN devices d ON d.id=e.device_id WHERE e.received_at >= $1 AND e.received_at < $2 ORDER BY e.received_at, e.id LIMIT $3`, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []RebuildEvent
	for rows.Next() {
		var record RebuildEvent
		if err := rows.Scan(&record.DeviceID, &record.Event.EventID, &record.Event.MeasurementType, &record.Event.Value, &record.Event.Unit, &record.Event.Timestamp, &record.ReceivedAt); err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

type RedisRebuilder struct {
	source    RebuildSource
	publisher *RedisStatePublisher
	mu        sync.Mutex
}

func NewRedisRebuilder(source RebuildSource, publisher *RedisStatePublisher) (*RedisRebuilder, error) {
	if source == nil || publisher == nil {
		return nil, errors.New("rebuild source and publisher are required")
	}
	return &RedisRebuilder{source: source, publisher: publisher}, nil
}

// Rebuild is deliberately operator-facing infrastructure, not a live-read
// fallback. It permits one short, capped scan at a time.
func (r *RedisRebuilder) Rebuild(ctx context.Context, from, to time.Time) (int, error) {
	if !from.Before(to) || to.Sub(from) > maxRebuildRange {
		return 0, errors.New("rebuild range must be positive and at most 15 minutes")
	}
	if !r.mu.TryLock() {
		return 0, ErrRebuildBusy
	}
	defer r.mu.Unlock()
	records, err := r.source.LoadForRebuild(ctx, from, to, maxRebuildEvents)
	if err != nil {
		return 0, err
	}
	now := from
	engine := NewAggregateEngine(func() time.Time { return now })
	for _, record := range records {
		now = record.ReceivedAt
		snapshots := engine.Process(ingestion.ValidatedBatch{Device: credentials.AuthorizedDevice{ExternalID: record.DeviceID}, Events: []ingestion.ValidatedEvent{record.Event}})
		r.publisher.Publish(record.DeviceID, now, snapshots)
	}
	return len(records), nil
}
