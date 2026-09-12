package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisStatePublisherPublishesVersionedExpiringSnapshotAndFreshness(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7.4.2-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start Redis: %v", err)
	}
	testcontainers.CleanupContainer(t, container)
	address, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: address})
	t.Cleanup(func() { _ = client.Close() })

	publisher, err := NewRedisStatePublisher(client, RedisPublisherConfig{
		FlushInterval: 5 * time.Millisecond,
		Timeout:       100 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := publisher.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown publisher: %v", err)
		}
	})

	processedAt := time.Now().UTC().Truncate(time.Second)
	snapshot := AggregateSnapshot{
		DeviceID:        "edge-01",
		MeasurementType: "temperature",
		LatestValue:     70,
		LatestTimestamp: processedAt.Add(-10 * time.Second),
		Count:           2,
		Sum:             138,
		Average:         69,
		Minimum:         68,
		Maximum:         70,
		WindowStart:     processedAt.Add(-299 * time.Second),
		WindowEnd:       processedAt,
		Status:          ThresholdWarning,
		UpdatedAt:       processedAt,
	}
	publisher.Publish("edge-01", processedAt, []AggregateSnapshot{snapshot})

	waitForRedisState(t, func() bool {
		return client.Exists(ctx, aggregateRedisKey("edge-01", "temperature")).Val() == 1
	})
	values, err := client.HGetAll(ctx, aggregateRedisKey("edge-01", "temperature")).Result()
	if err != nil {
		t.Fatal(err)
	}
	if values["latest_value"] != "70" || values["count"] != "2" || values["status"] != "warning" || values["window_end"] != processedAt.Format(time.RFC3339Nano) {
		t.Fatalf("aggregate values=%#v", values)
	}
	ttl, err := client.TTL(ctx, aggregateRedisKey("edge-01", "temperature")).Result()
	if err != nil || ttl <= 0 || ttl > redisStateTTL {
		t.Fatalf("aggregate ttl=%s, err=%v", ttl, err)
	}
	if !client.SIsMember(ctx, deviceMeasurementsRedisKey("edge-01"), "temperature").Val() {
		t.Fatal("measurement set is missing the published measurement")
	}
	measurementTTL, err := client.TTL(ctx, deviceMeasurementsRedisKey("edge-01")).Result()
	if err != nil || measurementTTL <= 0 || measurementTTL > redisStateTTL {
		t.Fatalf("measurement set ttl=%s, err=%v", measurementTTL, err)
	}
	lastSeen, err := client.ZScore(ctx, deviceLastSeenRedisKey, "edge-01").Result()
	if err != nil || lastSeen != float64(processedAt.Unix()) {
		t.Fatalf("last seen=%v, err=%v", lastSeen, err)
	}
}

func TestRedisStatePublisherCoalescesAndDoesNotBlockOnRedisFailure(t *testing.T) {
	failedClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = failedClient.Close() })
	publisher, err := NewRedisStatePublisher(failedClient, RedisPublisherConfig{
		FlushInterval: time.Millisecond,
		Timeout:       5 * time.Millisecond,
		QueueCapacity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = publisher.Shutdown(shutdownCtx)
	})

	started := time.Now()
	for range 10_000 {
		publisher.Publish("edge-01", time.Now().UTC(), []AggregateSnapshot{{DeviceID: "edge-01", MeasurementType: "temperature"}})
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("publication blocked worker for %s", elapsed)
	}
	waitForRedisState(t, func() bool { return publisher.Metrics().Errors > 0 })
	metrics := publisher.Metrics()
	if metrics.Errors == 0 || metrics.Dropped == 0 {
		t.Fatalf("metrics=%#v, want observable errors and bounded dropped publications", metrics)
	}
	if publisher.Mode() != CacheModeDegraded {
		t.Fatalf("mode=%q, want %q after Redis failure", publisher.Mode(), CacheModeDegraded)
	}
}

func waitForRedisState(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("Redis state condition was not reached")
		}
		time.Sleep(time.Millisecond)
	}
}
