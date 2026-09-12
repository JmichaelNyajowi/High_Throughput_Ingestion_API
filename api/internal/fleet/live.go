package fleet

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Reader struct {
	redis redis.UniversalClient
	mode  func() string
	now   func() time.Time
}

var ErrDeviceNotFound = errors.New("device live aggregate not found")

func New(redisClient redis.UniversalClient, mode func() string) *Reader {
	return &Reader{redis: redisClient, mode: mode, now: time.Now}
}
func (r *Reader) Fleet(ctx context.Context) ([]Device, error) {
	if r.mode() == "degraded" {
		return nil, errors.New("degraded")
	}
	ids, err := r.redis.ZRange(ctx, "telemetry:v1:device-last-seen", 0, -1).Result()
	if err != nil {
		return nil, err
	}
	out := make([]Device, 0, len(ids))
	for _, id := range ids {
		d, e := r.Device(ctx, id)
		if e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, nil
}

type Aggregate struct {
	MeasurementType string    `json:"measurement_type"`
	Unit            string    `json:"unit"`
	LatestValue     float64   `json:"latest_value"`
	LatestTimestamp time.Time `json:"latest_timestamp"`
	Count           uint64    `json:"count"`
	Sum             float64   `json:"sum"`
	Average         float64   `json:"average"`
	Minimum         float64   `json:"minimum"`
	Maximum         float64   `json:"maximum"`
	WindowStart     time.Time `json:"window_start"`
	WindowEnd       time.Time `json:"window_end"`
	ThresholdStatus string    `json:"threshold_status"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type Device struct {
	DeviceID        string      `json:"device_id"`
	Freshness       string      `json:"freshness"`
	LastProcessedAt time.Time   `json:"last_processed_at"`
	OverallStatus   string      `json:"overall_status"`
	Measurements    []Aggregate `json:"measurements"`
}

func (r *Reader) Device(ctx context.Context, id string) (Device, error) {
	if r.mode() == "degraded" {
		return Device{}, errors.New("degraded")
	}
	score, err := r.redis.ZScore(ctx, "telemetry:v1:device-last-seen", id).Result()
	if err == redis.Nil {
		return Device{}, ErrDeviceNotFound
	}
	if err != nil {
		return Device{}, err
	}
	d := Device{DeviceID: id, LastProcessedAt: time.Unix(int64(score), 0).UTC(), Freshness: "active", OverallStatus: "normal"}
	if r.now().Sub(d.LastProcessedAt) > 60*time.Second {
		d.Freshness = "stale"
	}
	ms, err := r.redis.SMembers(ctx, "telemetry:v1:device-measurements:"+id).Result()
	if err != nil && err != redis.Nil {
		return Device{}, err
	}
	for _, m := range ms {
		h, err := r.redis.HGetAll(ctx, "telemetry:v1:aggregate:"+id+":"+m).Result()
		if err != nil {
			return Device{}, err
		}
		if len(h) == 0 {
			continue
		}
		a := Aggregate{MeasurementType: m, Unit: measurementUnit(m), ThresholdStatus: h["status"]}
		a.LatestValue, _ = strconv.ParseFloat(h["latest_value"], 64)
		a.Count, _ = strconv.ParseUint(h["count"], 10, 64)
		a.Sum, _ = strconv.ParseFloat(h["sum"], 64)
		a.Average, _ = strconv.ParseFloat(h["average"], 64)
		a.Minimum, _ = strconv.ParseFloat(h["minimum"], 64)
		a.Maximum, _ = strconv.ParseFloat(h["maximum"], 64)
		a.LatestTimestamp, _ = time.Parse(time.RFC3339Nano, h["latest_timestamp"])
		a.WindowStart, _ = time.Parse(time.RFC3339Nano, h["window_start"])
		a.WindowEnd, _ = time.Parse(time.RFC3339Nano, h["window_end"])
		a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, h["updated_at"])
		d.Measurements = append(d.Measurements, a)
		if a.ThresholdStatus == "critical" {
			d.OverallStatus = "critical"
		} else if a.ThresholdStatus == "warning" && d.OverallStatus != "critical" {
			d.OverallStatus = "warning"
		}
	}
	return d, nil
}

func measurementUnit(measurementType string) string {
	switch measurementType {
	case "temperature":
		return "C"
	case "voltage":
		return "V"
	case "battery":
		return "%"
	case "pressure":
		return "kPa"
	default:
		return ""
	}
}
