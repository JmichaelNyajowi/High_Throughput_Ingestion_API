package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
)

const (
	RetryBufferCapacityEvents = 10_000
	RetryMaximumAttempts      = 3
	RetryInitialDelay         = 100 * time.Millisecond
	RetryMaximumDelay         = 5 * time.Second
	retryBufferFullReason     = "retry buffer capacity reached"
	retryExhaustedReason      = "persistence retries exhausted"
)

// RetryConfig controls the bounded persistence retry worker. Wait is injectable
// only to make the backoff boundary deterministic in tests.
type RetryConfig struct {
	CapacityEvents int
	Wait           func(context.Context, time.Duration) error
	Logger         *slog.Logger
}

// RetryMetrics exposes bounded retry state to the operations module. Prometheus
// export is introduced by Ticket 15.
type RetryMetrics struct {
	OccupancyEvents int
	Retries         uint64
	Exhausted       uint64
	Saturated       uint64
	Backpressured   bool
}

type retryEntry struct {
	batches []ingestion.ValidatedBatch
	events  int
}

// RetryBuffer owns failed PostgreSQL flushes. It has no Redis dependency, so a
// healthy cache cannot hide a PostgreSQL safety limit.
type RetryBuffer struct {
	repository EventRepository
	capacity   int
	wait       func(context.Context, time.Duration) error
	logger     *slog.Logger

	mu      sync.RWMutex
	entries []retryEntry
	metrics RetryMetrics
	closed  bool

	wake     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
	context  context.Context
	cancel   context.CancelFunc
}

func NewRetryBuffer(repository EventRepository, config RetryConfig) (*RetryBuffer, error) {
	if repository == nil {
		return nil, errors.New("repository is required")
	}
	if config.CapacityEvents == 0 {
		config.CapacityEvents = RetryBufferCapacityEvents
	}
	if config.CapacityEvents < 1 || config.CapacityEvents > RetryBufferCapacityEvents {
		return nil, errors.New("retry capacity must be between 1 and 10000 events")
	}
	if config.Wait == nil {
		config.Wait = waitForRetry
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	workerContext, cancel := context.WithCancel(context.Background())
	buffer := &RetryBuffer{
		repository: repository,
		capacity:   config.CapacityEvents,
		wait:       config.Wait,
		logger:     config.Logger,
		wake:       make(chan struct{}, 1),
		done:       make(chan struct{}),
		context:    workerContext,
		cancel:     cancel,
	}
	go buffer.run()
	return buffer, nil
}

// Enqueue transfers a failed flush into the retry buffer. A rejected entry is
// intentionally dropped only after an explicit loss log and backpressure state.
func (buffer *RetryBuffer) Enqueue(batches []ingestion.ValidatedBatch) bool {
	events := eventCount(batches)
	if events == 0 {
		return true
	}
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	if buffer.closed || buffer.metrics.Backpressured || events > buffer.capacity-buffer.metrics.OccupancyEvents {
		buffer.activateBackpressureLocked(retryBufferFullReason, events)
		return false
	}
	buffer.entries = append(buffer.entries, retryEntry{batches: batches, events: events})
	buffer.metrics.OccupancyEvents += events
	select {
	case buffer.wake <- struct{}{}:
	default:
	}
	return true
}

// AllowsAdmission is the Telemetry-to-Ingestion safety seam. It is intentionally
// independent of Redis availability.
func (buffer *RetryBuffer) AllowsAdmission() bool {
	buffer.mu.RLock()
	defer buffer.mu.RUnlock()
	return !buffer.metrics.Backpressured && !buffer.closed
}

func (buffer *RetryBuffer) Metrics() RetryMetrics {
	buffer.mu.RLock()
	defer buffer.mu.RUnlock()
	return buffer.metrics
}

func (buffer *RetryBuffer) Shutdown(ctx context.Context) error {
	buffer.stopOnce.Do(func() {
		buffer.mu.Lock()
		buffer.closed = true
		buffer.mu.Unlock()
		buffer.cancel()
	})
	select {
	case <-buffer.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (buffer *RetryBuffer) run() {
	defer close(buffer.done)
	for {
		entry, found := buffer.next()
		if !found {
			select {
			case <-buffer.wake:
				continue
			case <-buffer.context.Done():
				return
			}
		}
		buffer.retry(entry)
	}
}

func (buffer *RetryBuffer) next() (retryEntry, bool) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	if len(buffer.entries) == 0 {
		return retryEntry{}, false
	}
	entry := buffer.entries[0]
	buffer.entries = buffer.entries[1:]
	return entry, true
}

func (buffer *RetryBuffer) retry(entry retryEntry) {
	for attempt := 1; attempt <= RetryMaximumAttempts; attempt++ {
		if err := buffer.wait(buffer.context, retryDelay(attempt)); err != nil {
			return
		}
		_, err := buffer.repository.Flush(buffer.context, entry.batches)
		buffer.mu.Lock()
		buffer.metrics.Retries++
		buffer.mu.Unlock()
		if err == nil {
			buffer.complete(entry.events)
			return
		}
	}
	buffer.mu.Lock()
	buffer.metrics.Exhausted++
	buffer.metrics.OccupancyEvents -= entry.events
	buffer.activateBackpressureLocked(retryExhaustedReason, entry.events)
	buffer.mu.Unlock()
}

func (buffer *RetryBuffer) complete(events int) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	buffer.metrics.OccupancyEvents -= events
}

func (buffer *RetryBuffer) activateBackpressureLocked(reason string, events int) {
	if reason == retryBufferFullReason {
		buffer.metrics.Saturated++
	}
	buffer.metrics.Backpressured = true
	buffer.logger.Error("telemetry persistence data dropped; admission backpressure activated", "reason", reason, "event_count", events)
}

func retryDelay(attempt int) time.Duration {
	delay := RetryInitialDelay
	for index := 1; index < attempt && delay < RetryMaximumDelay; index++ {
		delay *= 2
	}
	if delay > RetryMaximumDelay {
		return RetryMaximumDelay
	}
	return delay
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func eventCount(batches []ingestion.ValidatedBatch) int {
	count := 0
	for _, batch := range batches {
		count += len(batch.Events)
	}
	return count
}
