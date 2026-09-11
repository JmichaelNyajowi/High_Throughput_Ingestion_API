// Package telemetry owns asynchronous processing of already-admitted telemetry.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/example/telemetry/api/internal/ingestion"
)

const (
	maxShardCount        = 64
	maxCapacityPerShard  = 1_024
	maxQueueBatches      = 1_024
	queueUnavailableCode = "unavailable"
)

// QueueConfig bounds the fixed worker pool and its per-shard backlog.
type QueueConfig struct {
	ShardCount       int
	CapacityPerShard int
}

func (config QueueConfig) Validate() error {
	if config.ShardCount < 1 || config.ShardCount > maxShardCount {
		return fmt.Errorf("shard count must be between 1 and %d", maxShardCount)
	}
	if config.CapacityPerShard < 1 || config.CapacityPerShard > maxCapacityPerShard {
		return fmt.Errorf("per-shard capacity must be between 1 and %d", maxCapacityPerShard)
	}
	if config.ShardCount*config.CapacityPerShard > maxQueueBatches {
		return fmt.Errorf("total queue capacity must not exceed %d batches", maxQueueBatches)
	}
	return nil
}

// Processor receives batches serially for each device. Different device shards
// process concurrently.
type Processor func(context.Context, ingestion.ValidatedBatch)

// AdmissionError lets the HTTP composition layer turn an admission failure into
// the published response without exposing queue internals.
type AdmissionError struct {
	StatusCode int
	Code       string
	Message    string
}

func (err AdmissionError) Error() string {
	return err.Message
}

func (err AdmissionError) AdmissionStatusCode() int { return err.StatusCode }
func (err AdmissionError) AdmissionCode() string    { return err.Code }

var (
	ErrQueueFull = AdmissionError{
		StatusCode: 503,
		Code:       queueUnavailableCode,
		Message:    "Admission queue is full",
	}
	ErrAdmissionStopped = AdmissionError{
		StatusCode: 503,
		Code:       queueUnavailableCode,
		Message:    "Admission is paused",
	}
	ErrInvalidBatch = errors.New("validated batch has no authenticated device ID")
)

// QueueMetrics provides bounded, in-process queue state for the runtime
// monitoring seam. Prometheus collection is introduced by Ticket 15.
type QueueMetrics struct {
	Capacity     int
	Depth        int
	ShardDepths  []int
	WorkerCount  int
	WorkerPanics uint64
}

// DeviceShardedQueue has one channel and one worker per shard. It is an
// in-memory, non-durable admission queue: successful admission is not a
// persistence acknowledgement.
type DeviceShardedQueue struct {
	processor Processor
	logger    *slog.Logger
	shards    []chan ingestion.ValidatedBatch

	mu        sync.Mutex
	accepting bool
	stopOnce  sync.Once
	done      chan struct{}
	workers   sync.WaitGroup
	panics    atomic.Uint64
}

func NewDeviceShardedQueue(config QueueConfig, processor Processor, logger *slog.Logger) (*DeviceShardedQueue, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if processor == nil {
		return nil, errors.New("queue processor is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	queue := &DeviceShardedQueue{
		processor: processor,
		logger:    logger,
		shards:    make([]chan ingestion.ValidatedBatch, config.ShardCount),
		accepting: true,
		done:      make(chan struct{}),
	}
	for index := range queue.shards {
		queue.shards[index] = make(chan ingestion.ValidatedBatch, config.CapacityPerShard)
		queue.workers.Add(1)
		go queue.runWorker(queue.shards[index])
	}
	go func() {
		queue.workers.Wait()
		close(queue.done)
	}()
	return queue, nil
}

// Admit places a whole validated batch in its authenticated device's shard
// without blocking the HTTP path. It never starts a goroutine per request.
func (queue *DeviceShardedQueue) Admit(batch ingestion.ValidatedBatch) error {
	deviceID := batch.Device.ExternalID
	if deviceID == "" {
		return ErrInvalidBatch
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()
	if !queue.accepting {
		return ErrAdmissionStopped
	}
	select {
	case queue.shards[queue.shardFor(deviceID)] <- batch:
		return nil
	default:
		return ErrQueueFull
	}
}

// Pause prevents future admission while permitting already-admitted work to
// drain during controlled shutdown.
func (queue *DeviceShardedQueue) Pause() {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.accepting = false
}

// Shutdown stops admission, closes every shard once, and waits only until the
// caller-provided graceful-drain deadline expires.
func (queue *DeviceShardedQueue) Shutdown(ctx context.Context) error {
	queue.stopOnce.Do(func() {
		queue.mu.Lock()
		queue.accepting = false
		for _, shard := range queue.shards {
			close(shard)
		}
		queue.mu.Unlock()
	})

	select {
	case <-queue.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (queue *DeviceShardedQueue) Metrics() QueueMetrics {
	metrics := QueueMetrics{
		ShardDepths: make([]int, len(queue.shards)),
		WorkerCount: len(queue.shards),
	}
	for index, shard := range queue.shards {
		metrics.Capacity += cap(shard)
		metrics.ShardDepths[index] = len(shard)
		metrics.Depth += metrics.ShardDepths[index]
	}
	metrics.WorkerPanics = queue.panics.Load()
	return metrics
}

func (queue *DeviceShardedQueue) shardFor(deviceID string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(deviceID))
	return int(hash.Sum32() % uint32(len(queue.shards)))
}

func (queue *DeviceShardedQueue) runWorker(shard <-chan ingestion.ValidatedBatch) {
	defer queue.workers.Done()
	for batch := range shard {
		queue.processBatch(batch)
	}
}

func (queue *DeviceShardedQueue) processBatch(batch ingestion.ValidatedBatch) {
	defer func() {
		if recovered := recover(); recovered != nil {
			queue.panics.Add(1)
			queue.logger.Error("telemetry worker panic recovered")
		}
	}()
	queue.processor(context.Background(), batch)
}
