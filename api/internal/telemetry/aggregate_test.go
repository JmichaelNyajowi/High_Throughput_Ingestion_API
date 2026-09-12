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
	if snapshot.Status != ThresholdNormal {
		t.Fatalf("status=%q, want %q", snapshot.Status, ThresholdNormal)
	}
}
