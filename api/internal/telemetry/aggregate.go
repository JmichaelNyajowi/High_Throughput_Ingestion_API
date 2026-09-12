package telemetry

import (
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
)

const aggregateBucketCount = 300

type ThresholdStatus string

const (
	ThresholdNormal   ThresholdStatus = "normal"
	ThresholdWarning  ThresholdStatus = "warning"
	ThresholdCritical ThresholdStatus = "critical"
)

// AggregateSnapshot is the derived processing-time state consumed by the
// cache-publication seam. It never represents a durable raw-event read.
type AggregateSnapshot struct {
	DeviceID        string
	MeasurementType string
	LatestValue     float64
	LatestTimestamp time.Time
	Count           uint64
	Sum             float64
	Average         float64
	Minimum         float64
	Maximum         float64
	WindowStart     time.Time
	WindowEnd       time.Time
	Status          ThresholdStatus
	UpdatedAt       time.Time
}

type aggregateKey struct {
	deviceID        string
	measurementType string
}

type aggregateBucket struct {
	second int64
	count  uint64
	sum    float64
	min    float64
	max    float64
}

type aggregateWindow struct {
	buckets         [aggregateBucketCount]aggregateBucket
	latestValue     float64
	latestTimestamp time.Time
	updatedAt       time.Time
	hasLatest       bool
}

// AggregateEngine owns bounded five-minute processing-time state. A single
// instance is safe for the fixed concurrent shard workers.
type AggregateEngine struct {
	now func() time.Time

	mu      sync.RWMutex
	windows map[aggregateKey]*aggregateWindow
}

func NewAggregateEngine(now func() time.Time) *AggregateEngine {
	if now == nil {
		now = time.Now
	}
	return &AggregateEngine{now: now, windows: make(map[aggregateKey]*aggregateWindow)}
}

// Process applies only events that are not more than 60 seconds old at worker
// processing time. Raw persistence is intentionally handled separately.
func (engine *AggregateEngine) Process(batch ingestion.ValidatedBatch) []AggregateSnapshot {
	processingNow := engine.now().UTC()
	processedAt := processingNow.Truncate(time.Second)
	engine.mu.Lock()
	defer engine.mu.Unlock()

	updated := make(map[aggregateKey]struct{})
	for _, event := range batch.Events {
		if event.SkipLiveAggregation || event.Timestamp.Before(processingNow.Add(-60*time.Second)) {
			continue
		}
		key := aggregateKey{deviceID: batch.Device.ExternalID, measurementType: event.MeasurementType}
		window := engine.windows[key]
		if window == nil {
			window = &aggregateWindow{}
			engine.windows[key] = window
		}
		window.add(processedAt.Unix(), event.Value, event.Timestamp, processingNow)
		updated[key] = struct{}{}
	}

	snapshots := make([]AggregateSnapshot, 0, len(updated))
	for key := range updated {
		if snapshot, found := engine.snapshotLocked(key, processedAt); found {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

func (engine *AggregateEngine) Snapshot(deviceID, measurementType string) (AggregateSnapshot, bool) {
	processedAt := engine.now().UTC().Truncate(time.Second)
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	return engine.snapshotLocked(aggregateKey{deviceID: deviceID, measurementType: measurementType}, processedAt)
}

func (window *aggregateWindow) add(second int64, value float64, eventTimestamp, processedAt time.Time) {
	bucket := &window.buckets[int(second%aggregateBucketCount)]
	if bucket.second != second {
		*bucket = aggregateBucket{second: second, min: value, max: value}
	}
	bucket.count++
	bucket.sum += value
	if value < bucket.min {
		bucket.min = value
	}
	if value > bucket.max {
		bucket.max = value
	}
	window.latestValue = value
	window.latestTimestamp = eventTimestamp
	window.updatedAt = processedAt
	window.hasLatest = true
}

func (engine *AggregateEngine) snapshotLocked(key aggregateKey, processedAt time.Time) (AggregateSnapshot, bool) {
	window := engine.windows[key]
	if window == nil || !window.hasLatest {
		return AggregateSnapshot{}, false
	}
	endSecond := processedAt.Unix()
	startSecond := endSecond - aggregateBucketCount + 1
	snapshot := AggregateSnapshot{
		DeviceID:        key.deviceID,
		MeasurementType: key.measurementType,
		LatestValue:     window.latestValue,
		LatestTimestamp: window.latestTimestamp,
		WindowStart:     time.Unix(startSecond, 0).UTC(),
		WindowEnd:       time.Unix(endSecond, 0).UTC(),
		UpdatedAt:       window.updatedAt,
	}
	for _, bucket := range window.buckets {
		if bucket.count == 0 || bucket.second < startSecond || bucket.second > endSecond {
			continue
		}
		snapshot.Count += bucket.count
		snapshot.Sum += bucket.sum
		if snapshot.Count == bucket.count || bucket.min < snapshot.Minimum {
			snapshot.Minimum = bucket.min
		}
		if snapshot.Count == bucket.count || bucket.max > snapshot.Maximum {
			snapshot.Maximum = bucket.max
		}
	}
	if snapshot.Count == 0 {
		return AggregateSnapshot{}, false
	}
	snapshot.Average = snapshot.Sum / float64(snapshot.Count)
	snapshot.Status = thresholdStatus(key.measurementType, snapshot.LatestValue)
	return snapshot, true
}

func thresholdStatus(measurementType string, value float64) ThresholdStatus {
	switch measurementType {
	case "temperature":
		if value > 85 {
			return ThresholdCritical
		}
		if value >= 70 {
			return ThresholdWarning
		}
	case "voltage":
		if value < 10.5 {
			return ThresholdCritical
		}
		if value < 11.5 {
			return ThresholdWarning
		}
	case "battery":
		if value < 10 {
			return ThresholdCritical
		}
		if value < 20 {
			return ThresholdWarning
		}
	case "pressure":
		if value > 700 {
			return ThresholdCritical
		}
		if value > 300 {
			return ThresholdWarning
		}
	}
	return ThresholdNormal
}
