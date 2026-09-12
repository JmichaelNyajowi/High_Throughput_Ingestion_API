package telemetry

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
)

func TestRetryBufferRetriesFailedFlushWithExponentialBackoff(t *testing.T) {
	repository := &scriptedRetryRepository{results: []error{errors.New("postgres unavailable"), errors.New("postgres unavailable"), nil}, completed: make(chan struct{})}
	var delays []time.Duration
	var delaysMu sync.Mutex
	buffer, err := NewRetryBuffer(repository, RetryConfig{
		CapacityEvents: RetryBufferCapacityEvents,
		Wait: func(_ context.Context, delay time.Duration) error {
			delaysMu.Lock()
			delays = append(delays, delay)
			delaysMu.Unlock()
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := buffer.Shutdown(shutdownContext); err != nil {
			t.Errorf("shutdown retry buffer: %v", err)
		}
	})

	if !buffer.Enqueue([]ingestion.ValidatedBatch{testPersistenceBatch(2)}) {
		t.Fatal("enqueue failed retry batch")
	}
	select {
	case <-repository.completed:
	case <-time.After(time.Second):
		t.Fatal("retry did not succeed")
	}

	delaysMu.Lock()
	defer delaysMu.Unlock()
	if len(delays) != 3 || delays[0] != 100*time.Millisecond || delays[1] != 200*time.Millisecond || delays[2] != 400*time.Millisecond {
		t.Fatalf("retry delays=%v, want [100ms 200ms 400ms]", delays)
	}
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != 0 || metrics.Retries != 3 || metrics.Exhausted != 0 {
		t.Fatalf("retry metrics=%#v", metrics)
	}
	if !buffer.AllowsAdmission() {
		t.Fatal("successful retry must not activate admission backpressure")
	}
}

func TestBatcherMovesFailedFlushIntoBoundedRetryBuffer(t *testing.T) {
	repository := &scriptedRetryRepository{results: []error{errors.New("postgres unavailable"), nil}, completed: make(chan struct{})}
	buffer, err := NewRetryBuffer(repository, RetryConfig{
		CapacityEvents: RetryBufferCapacityEvents,
		Wait:           func(context.Context, time.Duration) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	batcher, err := NewBatcher(repository, 1, buffer)
	if err != nil {
		t.Fatal(err)
	}
	if !batcher.Submit(testPersistenceBatch(2)) {
		t.Fatal("submit")
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := batcher.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
	select {
	case <-repository.completed:
	case <-time.After(time.Second):
		t.Fatal("failed flush was not retried")
	}
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != 0 || metrics.Retries != 1 || metrics.Backpressured {
		t.Fatalf("retry metrics=%#v", metrics)
	}
	if err := buffer.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
}

func TestBatcherProcessorMovesSaturatedInputIntoRetryBuffer(t *testing.T) {
	repository := &scriptedRetryRepository{results: []error{nil}, completed: make(chan struct{})}
	buffer, err := NewRetryBuffer(repository, RetryConfig{
		CapacityEvents: RetryBufferCapacityEvents,
		Wait:           func(context.Context, time.Duration) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	batcher := &Batcher{input: make(chan ingestion.ValidatedBatch, 1), retry: buffer}
	batcher.input <- testPersistenceBatch(1)

	batcher.Processor(context.Background(), testPersistenceBatch(2))

	select {
	case <-repository.completed:
	case <-time.After(time.Second):
		t.Fatal("saturated persistence input was not transferred to the retry buffer")
	}
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != 0 || metrics.Backpressured {
		t.Fatalf("retry metrics=%#v", metrics)
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := buffer.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
}

func TestRetryBufferExhaustionActivatesAdmissionBackpressure(t *testing.T) {
	repository := &scriptedRetryRepository{results: []error{errors.New("postgres unavailable"), errors.New("postgres unavailable"), errors.New("postgres unavailable")}, completed: make(chan struct{})}
	var logs bytes.Buffer
	buffer, err := NewRetryBuffer(repository, RetryConfig{
		CapacityEvents: RetryBufferCapacityEvents,
		Wait:           func(context.Context, time.Duration) error { return nil },
		Logger:         slog.New(slog.NewJSONHandler(&logs, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !buffer.Enqueue([]ingestion.ValidatedBatch{testPersistenceBatch(2)}) {
		t.Fatal("enqueue")
	}
	waitForRetryCondition(t, func() bool { return buffer.Metrics().Exhausted == 1 })
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != 0 || metrics.Retries != RetryMaximumAttempts || !metrics.Backpressured {
		t.Fatalf("retry metrics=%#v", metrics)
	}
	if buffer.AllowsAdmission() {
		t.Fatal("exhausted persistence retries must block admission")
	}
	if !bytes.Contains(logs.Bytes(), []byte(`"reason":"persistence retries exhausted"`)) || !bytes.Contains(logs.Bytes(), []byte(`"event_count":2`)) {
		t.Fatalf("loss log=%s, want structured exhaustion reason and event count", logs.String())
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := buffer.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
}

func TestRetryBufferCapacityIsTenThousandEventsAndObservable(t *testing.T) {
	firstWait := make(chan struct{})
	repository := &scriptedRetryRepository{results: []error{errors.New("postgres unavailable")}, completed: make(chan struct{})}
	buffer, err := NewRetryBuffer(repository, RetryConfig{
		CapacityEvents: RetryBufferCapacityEvents,
		Wait: func(ctx context.Context, _ time.Duration) error {
			select {
			case <-firstWait:
			default:
				close(firstWait)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !buffer.Enqueue([]ingestion.ValidatedBatch{testPersistenceBatch(RetryBufferCapacityEvents)}) {
		t.Fatal("fill retry buffer")
	}
	select {
	case <-firstWait:
	case <-time.After(time.Second):
		t.Fatal("retry worker did not start")
	}
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != RetryBufferCapacityEvents || metrics.Backpressured {
		t.Fatalf("filled retry metrics=%#v", metrics)
	}
	if buffer.Enqueue([]ingestion.ValidatedBatch{testPersistenceBatch(1)}) {
		t.Fatal("retry buffer accepted an event over its capacity")
	}
	if metrics := buffer.Metrics(); metrics.OccupancyEvents != RetryBufferCapacityEvents || metrics.Saturated != 1 || !metrics.Backpressured {
		t.Fatalf("saturated retry metrics=%#v", metrics)
	}
	if buffer.AllowsAdmission() {
		t.Fatal("saturated retry buffer must block admission")
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := buffer.Shutdown(shutdownContext); err != nil {
		t.Fatal(err)
	}
}

func TestRetryDelayIsExponentialAndCapped(t *testing.T) {
	if retryDelay(1) != 100*time.Millisecond || retryDelay(2) != 200*time.Millisecond || retryDelay(3) != 400*time.Millisecond || retryDelay(10) != RetryMaximumDelay {
		t.Fatalf("retry delay schedule = [%s %s %s %s], want 100ms, 200ms, 400ms, and 5s cap", retryDelay(1), retryDelay(2), retryDelay(3), retryDelay(10))
	}
}

func waitForRetryCondition(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("retry condition was not reached")
		}
		time.Sleep(time.Millisecond)
	}
}

type scriptedRetryRepository struct {
	mu        sync.Mutex
	results   []error
	completed chan struct{}
}

func (repository *scriptedRetryRepository) Flush(_ context.Context, batches []ingestion.ValidatedBatch) (FlushOutcome, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	count := 0
	for _, batch := range batches {
		count += len(batch.Events)
	}
	result := repository.results[0]
	repository.results = repository.results[1:]
	if result == nil {
		close(repository.completed)
		return FlushOutcome{Attempted: count, Inserted: count}, nil
	}
	return FlushOutcome{Attempted: count}, result
}
