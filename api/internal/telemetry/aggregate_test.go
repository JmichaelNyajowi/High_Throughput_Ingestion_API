package telemetry

import (
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion"
	"github.com/example/telemetry/api/internal/ingestion/credentials"
)

func TestAggregateEngineBuildsFiveMinuteProcessingTimeSnapshot(t *testing.T) {
	now := time.Date(2026, time.September, 12, 10, 0, 30, 0, time.UTC)
	engine := NewAggregateEngine(func() time.Time { return now })
	firstTimestamp := now.Add(-10 * time.Second)
	secondTimestamp := now.Add(-5 * time.Second)
	engine.Process(ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: "edge-01"},
		Events: []ingestion.ValidatedEvent{
			{MeasurementType: "temperature", Value: 68, Timestamp: firstTimestamp},
			{MeasurementType: "temperature", Value: 70, Timestamp: secondTimestamp},
		},
	})

	snapshot, found := engine.Snapshot("edge-01", "temperature")
	if !found {
		t.Fatal("temperature aggregate was not created")
	}
	if snapshot.Count != 2 || snapshot.Sum != 138 || snapshot.Average != 69 || snapshot.Minimum != 68 || snapshot.Maximum != 70 || snapshot.LatestValue != 70 || !snapshot.LatestTimestamp.Equal(secondTimestamp) {
		t.Fatalf("snapshot=%#v", snapshot)
	}
	if want := now.Truncate(time.Second).Add(-299 * time.Second); !snapshot.WindowStart.Equal(want) {
		t.Fatalf("window start=%s, want %s", snapshot.WindowStart, want)
	}
	if want := now.Truncate(time.Second); !snapshot.WindowEnd.Equal(want) {
		t.Fatalf("window end=%s, want %s", snapshot.WindowEnd, want)
	}
	if snapshot.Status != ThresholdWarning {
		t.Fatalf("status=%q, want %q", snapshot.Status, ThresholdWarning)
	}
}

func TestAggregateEngineSkipsEventsLateAtWorkerProcessingTime(t *testing.T) {
	now := time.Date(2026, time.September, 12, 10, 0, 30, 0, time.UTC)
	engine := NewAggregateEngine(func() time.Time { return now })
	engine.Process(ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: "edge-01"},
		Events: []ingestion.ValidatedEvent{{
			MeasurementType: "voltage",
			Value:           12,
			Timestamp:       now.Add(-61 * time.Second),
		}},
	})

	if _, found := engine.Snapshot("edge-01", "voltage"); found {
		t.Fatal("event more than 60 seconds late must not create a live aggregate")
	}
}

func TestAggregateEngineAcceptsEventAtLateEventBoundary(t *testing.T) {
	now := time.Date(2026, time.September, 12, 10, 0, 30, 900_000_000, time.UTC)
	engine := NewAggregateEngine(func() time.Time { return now })
	engine.Process(ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: "edge-01"},
		Events: []ingestion.ValidatedEvent{{
			MeasurementType: "voltage",
			Value:           12,
			Timestamp:       now.Add(-60 * time.Second),
		}},
	})

	if _, found := engine.Snapshot("edge-01", "voltage"); !found {
		t.Fatal("event exactly 60 seconds old must remain eligible for live aggregation")
	}
}

func TestAggregateEngineSkipsLateEventWithSubsecondWorkerClock(t *testing.T) {
	now := time.Date(2026, time.September, 12, 10, 0, 30, 900_000_000, time.UTC)
	engine := NewAggregateEngine(func() time.Time { return now })
	engine.Process(ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: "edge-01"},
		Events: []ingestion.ValidatedEvent{{
			MeasurementType: "voltage",
			Value:           12,
			Timestamp:       now.Add(-60*time.Second - time.Nanosecond),
		}},
	})

	if _, found := engine.Snapshot("edge-01", "voltage"); found {
		t.Fatal("event even fractionally more than 60 seconds old must not mutate live state")
	}
}

func TestAggregateEngineExpiresBucketsOutsideFiveMinuteWindow(t *testing.T) {
	now := time.Date(2026, time.September, 12, 10, 0, 0, 0, time.UTC)
	engine := NewAggregateEngine(func() time.Time { return now })
	engine.Process(ingestion.ValidatedBatch{
		Device: credentials.AuthorizedDevice{ExternalID: "edge-01"},
		Events: []ingestion.ValidatedEvent{{
			MeasurementType: "pressure",
			Value:           300,
			Timestamp:       now,
		}},
	})
	now = now.Add(5 * time.Minute)

	if _, found := engine.Snapshot("edge-01", "pressure"); found {
		t.Fatal("bucket older than five minutes must leave the live window")
	}
}

func TestThresholdStatusUsesExactPRDBoundaries(t *testing.T) {
	for _, test := range []struct {
		measurement string
		value       float64
		want        ThresholdStatus
	}{
		{"temperature", 69.999, ThresholdNormal},
		{"temperature", 70, ThresholdWarning},
		{"temperature", 85, ThresholdWarning},
		{"temperature", 85.001, ThresholdCritical},
		{"voltage", 11.5, ThresholdNormal},
		{"voltage", 10.5, ThresholdWarning},
		{"voltage", 11.499, ThresholdWarning},
		{"voltage", 10.499, ThresholdCritical},
		{"battery", 20, ThresholdNormal},
		{"battery", 10, ThresholdWarning},
		{"battery", 19.999, ThresholdWarning},
		{"battery", 9.999, ThresholdCritical},
		{"pressure", 300, ThresholdNormal},
		{"pressure", 300.001, ThresholdWarning},
		{"pressure", 700, ThresholdWarning},
		{"pressure", 700.001, ThresholdCritical},
	} {
		if got := thresholdStatus(test.measurement, test.value); got != test.want {
			t.Errorf("%s at %v = %q, want %q", test.measurement, test.value, got, test.want)
		}
	}
}
