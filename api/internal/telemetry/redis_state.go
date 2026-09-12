package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisStateTTL             = 15 * time.Minute
	defaultRedisFlushInterval = 250 * time.Millisecond
	defaultRedisTimeout       = 100 * time.Millisecond
	defaultRedisQueueCapacity = 1_024
	deviceLastSeenRedisKey    = "telemetry:v1:device-last-seen"
)

// RedisPublisherConfig keeps cache publication bounded and independent from
// the persistence path. Redis is derived state, so a full publication queue is
// observable loss, not a reason to block a shard worker.
type RedisPublisherConfig struct {
	FlushInterval time.Duration
	Timeout       time.Duration
	QueueCapacity int
	Logger        *slog.Logger
}

type RedisPublicationMetrics struct {
	Published uint64
	Errors    uint64
	Dropped   uint64
}

type CacheMode string

const (
	CacheModeLive     CacheMode = "live"
	CacheModeDegraded CacheMode = "degraded"
)

type redisPublication struct {
	deviceID    string
	processedAt time.Time
	snapshots   []AggregateSnapshot
}

// RedisStatePublisher coalesces compact aggregate snapshots before writing the
// versioned live-state key model. It owns no durable data and never performs
// Redis I/O on the worker call path.
type RedisStatePublisher struct {
	client       redis.UniversalClient
	timeout      time.Duration
	logger       *slog.Logger
	publications chan redisPublication
	done         chan struct{}
	stopOnce     sync.Once
	mu           sync.RWMutex
	closed       bool
	published    atomic.Uint64
	errors       atomic.Uint64
	dropped      atomic.Uint64
	mode         atomic.Int32
}

func NewRedisStatePublisher(client redis.UniversalClient, config RedisPublisherConfig) (*RedisStatePublisher, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = defaultRedisFlushInterval
	}
	if config.FlushInterval <= 0 || config.FlushInterval > defaultRedisFlushInterval {
		return nil, fmt.Errorf("redis flush interval must be between 1ns and %s", defaultRedisFlushInterval)
	}
	if config.Timeout == 0 {
		config.Timeout = defaultRedisTimeout
	}
	if config.Timeout <= 0 || config.Timeout > defaultRedisFlushInterval {
		return nil, fmt.Errorf("redis timeout must be between 1ns and %s", defaultRedisFlushInterval)
	}
	if config.QueueCapacity == 0 {
		config.QueueCapacity = defaultRedisQueueCapacity
	}
	if config.QueueCapacity < 1 || config.QueueCapacity > defaultRedisQueueCapacity {
		return nil, fmt.Errorf("redis publication queue capacity must be between 1 and %d", defaultRedisQueueCapacity)
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	publisher := &RedisStatePublisher{
		client:       client,
		timeout:      config.Timeout,
		logger:       config.Logger,
		publications: make(chan redisPublication, config.QueueCapacity),
		done:         make(chan struct{}),
	}
	go publisher.run(config.FlushInterval)
	return publisher, nil
}

// Publish transfers derived state to the bounded cache-publication queue. It
// deliberately does not wait for a Redis connection or write completion.
func (publisher *RedisStatePublisher) Publish(deviceID string, processedAt time.Time, snapshots []AggregateSnapshot) {
	publication := redisPublication{
		deviceID:    deviceID,
		processedAt: processedAt.UTC(),
		snapshots:   append([]AggregateSnapshot(nil), snapshots...),
	}
	publisher.mu.RLock()
	defer publisher.mu.RUnlock()
	if publisher.closed {
		publisher.dropped.Add(1)
		return
	}
	select {
	case publisher.publications <- publication:
	default:
		publisher.dropped.Add(1)
	}
}

func (publisher *RedisStatePublisher) Shutdown(ctx context.Context) error {
	publisher.stopOnce.Do(func() {
		publisher.mu.Lock()
		publisher.closed = true
		close(publisher.publications)
		publisher.mu.Unlock()
	})
	select {
	case <-publisher.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (publisher *RedisStatePublisher) Metrics() RedisPublicationMetrics {
	return RedisPublicationMetrics{
		Published: publisher.published.Load(),
		Errors:    publisher.errors.Load(),
		Dropped:   publisher.dropped.Load(),
	}
}

// Mode exposes cache availability only. Consumers must never substitute a
// PostgreSQL query when it is degraded.
func (publisher *RedisStatePublisher) Mode() CacheMode {
	if publisher.mode.Load() == 1 {
		return CacheModeDegraded
	}
	return CacheModeLive
}

func (publisher *RedisStatePublisher) run(interval time.Duration) {
	defer close(publisher.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	dirty := make(map[aggregateKey]AggregateSnapshot)
	lastSeen := make(map[string]time.Time)
	flush := func() {
		if len(dirty) == 0 && len(lastSeen) == 0 {
			return
		}
		if err := publisher.flush(dirty, lastSeen); err != nil {
			publisher.errors.Add(1)
			if publisher.mode.CompareAndSwap(0, 1) {
				publisher.logger.Warn("redis cache mode changed", "mode", CacheModeDegraded, "error", err)
			}
			publisher.logger.Warn("redis live-state publication failed", "error", err)
		} else {
			publisher.published.Add(uint64(len(dirty)))
			if publisher.mode.CompareAndSwap(1, 0) {
				publisher.logger.Info("redis cache mode changed", "mode", CacheModeLive)
			}
		}
		clear(dirty)
		clear(lastSeen)
	}
	for {
		select {
		case publication, open := <-publisher.publications:
			if !open {
				flush()
				return
			}
			if prior, found := lastSeen[publication.deviceID]; !found || publication.processedAt.After(prior) {
				lastSeen[publication.deviceID] = publication.processedAt
			}
			for _, snapshot := range publication.snapshots {
				dirty[aggregateKey{deviceID: snapshot.DeviceID, measurementType: snapshot.MeasurementType}] = snapshot
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (publisher *RedisStatePublisher) flush(dirty map[aggregateKey]AggregateSnapshot, lastSeen map[string]time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), publisher.timeout)
	defer cancel()
	pipeline := publisher.client.Pipeline()
	for _, snapshot := range dirty {
		key := aggregateRedisKey(snapshot.DeviceID, snapshot.MeasurementType)
		pipeline.HSet(ctx, key, aggregateRedisValues(snapshot))
		pipeline.Expire(ctx, key, redisStateTTL)
		measurementKey := deviceMeasurementsRedisKey(snapshot.DeviceID)
		pipeline.SAdd(ctx, measurementKey, snapshot.MeasurementType)
		pipeline.Expire(ctx, measurementKey, redisStateTTL)
	}
	for deviceID, processedAt := range lastSeen {
		pipeline.ZAdd(ctx, deviceLastSeenRedisKey, redis.Z{Score: float64(processedAt.Unix()), Member: deviceID})
	}
	pipeline.ZRemRangeByScore(ctx, deviceLastSeenRedisKey, "-inf", strconv.FormatInt(time.Now().UTC().Add(-redisStateTTL).Unix(), 10))
	_, err := pipeline.Exec(ctx)
	return err
}

func aggregateRedisKey(deviceID, measurementType string) string {
	return "telemetry:v1:aggregate:" + deviceID + ":" + measurementType
}

func deviceMeasurementsRedisKey(deviceID string) string {
	return "telemetry:v1:device-measurements:" + deviceID
}

func aggregateRedisValues(snapshot AggregateSnapshot) map[string]any {
	return map[string]any{
		"latest_value":     strconv.FormatFloat(snapshot.LatestValue, 'g', -1, 64),
		"latest_timestamp": snapshot.LatestTimestamp.UTC().Format(time.RFC3339Nano),
		"count":            strconv.FormatUint(snapshot.Count, 10),
		"sum":              strconv.FormatFloat(snapshot.Sum, 'g', -1, 64),
		"average":          strconv.FormatFloat(snapshot.Average, 'g', -1, 64),
		"minimum":          strconv.FormatFloat(snapshot.Minimum, 'g', -1, 64),
		"maximum":          strconv.FormatFloat(snapshot.Maximum, 'g', -1, 64),
		"window_start":     snapshot.WindowStart.UTC().Format(time.RFC3339Nano),
		"window_end":       snapshot.WindowEnd.UTC().Format(time.RFC3339Nano),
		"status":           string(snapshot.Status),
		"updated_at":       snapshot.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
