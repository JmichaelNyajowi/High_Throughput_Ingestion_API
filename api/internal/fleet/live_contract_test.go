package fleet

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAggregateJSONMatchesLiveContract(t *testing.T) {
	snapshot := Aggregate{
		MeasurementType: "temperature", Unit: "C", LatestValue: 76.2, LatestTimestamp: time.Now().UTC(), Count: 3,
		Sum: 216.3, Average: 72.1, Minimum: 68.4, Maximum: 76.2, WindowStart: time.Now().UTC(), WindowEnd: time.Now().UTC(), ThresholdStatus: "warning", UpdatedAt: time.Now().UTC(),
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"measurement_type", "unit", "latest_value", "latest_timestamp", "count", "sum", "average", "minimum", "maximum", "window_start", "window_end", "threshold_status", "updated_at"} {
		if _, ok := body[field]; !ok {
			t.Errorf("missing contract field %q in %s", field, encoded)
		}
	}
	if _, ok := body["Average"]; ok {
		t.Error("Go field name leaked into JSON contract")
	}
}
