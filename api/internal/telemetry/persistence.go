package telemetry

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PersistenceFlushSize  = 1_000
	PersistenceFlushEvery = 2 * time.Second
)

type FlushOutcome struct{ Attempted, Inserted, Conflicts int }

type EventRepository interface {
	Flush(context.Context, []ingestion.ValidatedBatch) (FlushOutcome, error)
}

type PostgresEventRepository struct{ pool *pgxpool.Pool }

func NewPostgresEventRepository(pool *pgxpool.Pool) PostgresEventRepository {
	return PostgresEventRepository{pool: pool}
}

func (repository PostgresEventRepository) Flush(ctx context.Context, batches []ingestion.ValidatedBatch) (FlushOutcome, error) {
	var outcome FlushOutcome
	for _, telemetryBatch := range batches {
		outcome.Attempted += len(telemetryBatch.Events)
	}
	if repository.pool == nil {
		return outcome, errors.New("postgres pool is required")
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return outcome, err
	}
	defer tx.Rollback(ctx)
	batch := &pgx.Batch{}
	for _, telemetryBatch := range batches {
		for _, event := range telemetryBatch.Events {
			batch.Queue(`INSERT INTO telemetry_events (event_id, device_id, measurement_type, value, unit, event_timestamp, received_at, request_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (device_id, event_id) DO NOTHING`, event.EventID, telemetryBatch.Device.InternalID, event.MeasurementType, event.Value, event.Unit, event.Timestamp, time.Now().UTC(), telemetryBatch.RequestID)
		}
	}
	results := tx.SendBatch(ctx, batch)
	for range outcome.Attempted {
		command, err := results.Exec()
		if err != nil {
			return outcome, err
		}
		outcome.Inserted += int(command.RowsAffected())
	}
	if err := results.Close(); err != nil {
		return outcome, err
	}
	outcome.Conflicts = outcome.Attempted - outcome.Inserted
	if err := tx.Commit(ctx); err != nil {
		return outcome, err
	}
	return outcome, nil
}

type PersistenceMetrics struct {
	LastFlush         FlushOutcome
	LastFlushDuration time.Duration
	Flushes           uint64
	Failures          uint64
}

type Batcher struct {
	repository EventRepository
	retry      *RetryBuffer
	input      chan ingestion.ValidatedBatch
	done       chan struct{}
	stopOnce   sync.Once
	mu         sync.RWMutex
	closed     bool
	metrics    PersistenceMetrics
}

func NewBatcher(repository EventRepository, capacity int, retries ...*RetryBuffer) (*Batcher, error) {
	if repository == nil || capacity < 1 {
		return nil, errors.New("repository and positive capacity are required")
	}
	b := &Batcher{repository: repository, retry: firstRetryBuffer(retries), input: make(chan ingestion.ValidatedBatch, capacity), done: make(chan struct{})}
	go b.run()
	return b, nil
}
func (b *Batcher) Submit(batch ingestion.ValidatedBatch) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return false
	}
	select {
	case b.input <- batch:
		return true
	default:
		return false
	}
}
func (b *Batcher) Shutdown(ctx context.Context) error {
	b.stopOnce.Do(func() {
		b.mu.Lock()
		b.closed = true
		close(b.input)
		b.mu.Unlock()
	})
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (b *Batcher) run() {
	defer close(b.done)
	ticker := time.NewTicker(PersistenceFlushEvery)
	defer ticker.Stop()
	pending := make([]ingestion.ValidatedBatch, 0)
	count := 0
	flush := func() {
		if count == 0 {
			return
		}
		started := time.Now()
		outcome, err := b.repository.Flush(context.Background(), pending)
		b.mu.Lock()
		b.metrics.LastFlush = outcome
		b.metrics.LastFlushDuration = time.Since(started)
		b.metrics.Flushes++
		if err != nil {
			b.metrics.Failures++
		}
		b.mu.Unlock()
		if err != nil {
			if b.retry != nil {
				b.retry.Enqueue(pending)
				pending = nil
				count = 0
			}
			return
		}
		pending = nil
		count = 0
	}
	for {
		select {
		case batch, ok := <-b.input:
			if !ok {
				flush()
				return
			}
			pending = append(pending, batch)
			count += len(batch.Events)
			if count >= PersistenceFlushSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func firstRetryBuffer(retries []*RetryBuffer) *RetryBuffer {
	if len(retries) == 0 {
		return nil
	}
	return retries[0]
}

func (b *Batcher) Metrics() PersistenceMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.metrics
}

// Processor is the fixed shard-worker callback. A saturated persistence input
// transfers the already-admitted batch to the bounded retry buffer instead of
// silently discarding it. If that buffer is full, Enqueue logs the loss and
// activates admission backpressure.
func (b *Batcher) Processor(_ context.Context, batch ingestion.ValidatedBatch) {
	if b.Submit(batch) || b.retry == nil {
		return
	}
	_ = b.retry.Enqueue([]ingestion.ValidatedBatch{batch})
}
