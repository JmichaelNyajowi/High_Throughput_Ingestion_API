package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
)

func TestBatcherFlushesAtSizeAndShutdownIsIdempotent(t *testing.T) {
	repository := &recordingRepository{flushed: make(chan int, 1)}
	batcher, err := NewBatcher(repository, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !batcher.Submit(testPersistenceBatch(PersistenceFlushSize)) {
		t.Fatal("submit")
	}
	select {
	case count := <-repository.flushed:
		if count != PersistenceFlushSize {
			t.Fatalf("count=%d", count)
		}
	case <-time.After(time.Second):
		t.Fatal("size flush timed out")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := batcher.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := batcher.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if metrics := batcher.Metrics(); metrics.Flushes != 1 || metrics.LastFlush.Inserted != PersistenceFlushSize {
		t.Fatalf("metrics=%#v", metrics)
	}
}

func TestBatcherFlushesOnTimer(t *testing.T) {
	repository := &recordingRepository{flushed: make(chan int, 1)}
	batcher, err := NewBatcher(repository, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !batcher.Submit(testPersistenceBatch(1)) {
		t.Fatal("submit")
	}
	select {
	case count := <-repository.flushed:
		if count != 1 {
			t.Fatalf("count=%d", count)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timer flush timed out")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := batcher.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

type recordingRepository struct {
	mu      sync.Mutex
	flushed chan int
}

func (r *recordingRepository) Flush(_ context.Context, batches []ingestion.ValidatedBatch) (FlushOutcome, error) {
	count := 0
	for _, b := range batches {
		count += len(b.Events)
	}
	r.mu.Lock()
	r.flushed <- count
	r.mu.Unlock()
	return FlushOutcome{Attempted: count, Inserted: count}, nil
}
func testPersistenceBatch(count int) ingestion.ValidatedBatch {
	return ingestion.ValidatedBatch{Events: make([]ingestion.ValidatedEvent, count)}
}
