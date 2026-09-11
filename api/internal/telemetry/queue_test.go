package telemetry

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/ingestion/credentials"
)

func TestDeviceShardedQueueProcessesDeviceBatchesInAdmissionOrder(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	processed := make(chan string, 2)

	queue, err := NewDeviceShardedQueue(QueueConfig{ShardCount: 2, CapacityPerShard: 2}, func(_ context.Context, batch ingestion.ValidatedBatch) {
		eventID := batch.Events[0].EventID
		if eventID == "first" {
			close(firstStarted)
			<-releaseFirst
		}
		processed <- eventID
	}, nil)
	if err != nil {
		t.Fatalf("NewDeviceShardedQueue() error = %v", err)
	}
	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := queue.Shutdown(shutdownContext); err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
	})

	if err := queue.Admit(testQueueBatch("edge-01", "first")); err != nil {
		t.Fatalf("admit first batch: %v", err)
	}
	<-firstStarted
	if err := queue.Admit(testQueueBatch("edge-01", "second")); err != nil {
		t.Fatalf("admit second batch: %v", err)
	}
	close(releaseFirst)

	for _, want := range []string{"first", "second"} {
		select {
		case got := <-processed:
			if got != want {
				t.Fatalf("processed event = %q, want %q", got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %q batch", want)
		}
	}
}

func TestDeviceShardedQueueRejectsFullShardWithTypedUnavailableErrorAndMetrics(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	queue, err := NewDeviceShardedQueue(QueueConfig{ShardCount: 1, CapacityPerShard: 1}, func(_ context.Context, batch ingestion.ValidatedBatch) {
		if batch.Events[0].EventID == "first" {
			close(firstStarted)
			<-releaseFirst
		}
	}, nil)
	if err != nil {
		t.Fatalf("NewDeviceShardedQueue() error = %v", err)
	}

	if err := queue.Admit(testQueueBatch("edge-01", "first")); err != nil {
		t.Fatalf("admit first batch: %v", err)
	}
	<-firstStarted
	if err := queue.Admit(testQueueBatch("edge-01", "second")); err != nil {
		t.Fatalf("admit queued batch: %v", err)
	}
	if err := queue.Admit(testQueueBatch("edge-01", "third")); err == nil {
		t.Fatal("admit full shard error = nil, want typed 503")
	} else {
		var admissionError AdmissionError
		if !errors.As(err, &admissionError) || admissionError.StatusCode != 503 || admissionError.Code != "unavailable" {
			t.Fatalf("full shard error = %#v, want typed unavailable 503", err)
		}
	}
	if metrics := queue.Metrics(); metrics.Capacity != 1 || metrics.Depth != 1 || metrics.ShardDepths[0] != 1 || metrics.WorkerCount != 1 {
		t.Fatalf("queue metrics = %#v, want one bounded queued batch", metrics)
	}

	close(releaseFirst)
	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := queue.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestDeviceShardedQueueShutdownPausesAdmissionAndHonoursDrainDeadline(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	queue, err := NewDeviceShardedQueue(QueueConfig{ShardCount: 1, CapacityPerShard: 1}, func(_ context.Context, _ ingestion.ValidatedBatch) {
		close(started)
		<-release
	}, nil)
	if err != nil {
		t.Fatalf("NewDeviceShardedQueue() error = %v", err)
	}
	if err := queue.Admit(testQueueBatch("edge-01", "first")); err != nil {
		t.Fatalf("admit batch: %v", err)
	}
	<-started

	drainContext, cancelDrain := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDrain()
	if err := queue.Shutdown(drainContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want drain deadline", err)
	}
	if err := queue.Admit(testQueueBatch("edge-01", "after-pause")); !errors.Is(err, ErrAdmissionStopped) {
		t.Fatalf("admit after shutdown starts = %v, want paused admission", err)
	}

	close(release)
	completedContext, cancelCompleted := context.WithTimeout(context.Background(), time.Second)
	defer cancelCompleted()
	if err := queue.Shutdown(completedContext); err != nil {
		t.Fatalf("Shutdown() after drain error = %v", err)
	}
}

func TestQueueConfigRejectsUnboundedWorkerAndCapacityValues(t *testing.T) {
	for _, config := range []QueueConfig{
		{ShardCount: 0, CapacityPerShard: 1},
		{ShardCount: 1, CapacityPerShard: 0},
		{ShardCount: maxShardCount + 1, CapacityPerShard: 1},
		{ShardCount: 1, CapacityPerShard: maxCapacityPerShard + 1},
		{ShardCount: 16, CapacityPerShard: 65},
	} {
		if err := config.Validate(); err == nil {
			t.Fatalf("Validate(%#v) error = nil, want bounded configuration rejection", config)
		}
	}
}

func TestDeviceShardedQueueRecoversWorkerPanicWithoutLosingWorkerCapacity(t *testing.T) {
	var calls atomic.Int32
	processed := make(chan string, 1)
	queue, err := NewDeviceShardedQueue(QueueConfig{ShardCount: 1, CapacityPerShard: 2}, func(_ context.Context, batch ingestion.ValidatedBatch) {
		if calls.Add(1) == 1 {
			panic("processor bug")
		}
		processed <- batch.Events[0].EventID
	}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewDeviceShardedQueue() error = %v", err)
	}

	if err := queue.Admit(testQueueBatch("edge-01", "panic")); err != nil {
		t.Fatalf("admit panic batch: %v", err)
	}
	if err := queue.Admit(testQueueBatch("edge-01", "after-panic")); err != nil {
		t.Fatalf("admit post-panic batch: %v", err)
	}
	select {
	case got := <-processed:
		if got != "after-panic" {
			t.Fatalf("processed event = %q, want post-panic batch", got)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not retain capacity after processor panic")
	}
	if metrics := queue.Metrics(); metrics.WorkerPanics != 1 || metrics.WorkerCount != 1 {
		t.Fatalf("queue metrics after panic = %#v, want one recovered panic and one worker", metrics)
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := queue.Shutdown(shutdownContext); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func testQueueBatch(deviceID, eventID string) ingestion.ValidatedBatch {
	return ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: deviceID},
		Events: []ingestion.ValidatedEvent{{EventID: eventID, DeviceID: deviceID}},
	}
}
