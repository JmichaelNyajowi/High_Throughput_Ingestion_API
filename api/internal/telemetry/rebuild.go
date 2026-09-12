package telemetry

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/ingestion/credentials"
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
// Its PostgreSQL implementation belongs to the persistence seam added with
// the next integration tranche; this interface prevents live reads from ever
// substituting an unbounded source query.
type RebuildSource interface {
	LoadForRebuild(context.Context, time.Time, time.Time, int) ([]RebuildEvent, error)
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
