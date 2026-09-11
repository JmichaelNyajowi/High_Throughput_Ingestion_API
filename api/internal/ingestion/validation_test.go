package ingestion

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/httpserver"
)

func TestBatchValidatorRejectsWholeInvalidBatchBeforeDownstreamAdmission(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	validator := NewBatchValidator(func() time.Time { return now })

	for _, testCase := range []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
	}{
		{"non JSON", "text/plain", "not json", http.StatusUnsupportedMediaType},
		{"body exceeds one megabyte", "application/json", strings.Repeat("x", maxTelemetryBodyBytes+1), http.StatusRequestEntityTooLarge},
		{"empty events", "application/json", `{"events":[]}`, http.StatusBadRequest},
		{"over 500 events", "application/json", batchJSON(501, "edge-01", "temperature", "C", "20", now), http.StatusBadRequest},
		{"malformed JSON", "application/json", `{"events":[`, http.StatusBadRequest},
		{"unknown field", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"temperature","value":20,"unexpected":true}]}`, http.StatusBadRequest},
		{"missing required value", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"voltage"}]}`, http.StatusBadRequest},
		{"non finite value", "application/json", eventBatchJSON("edge-01", "temperature", "C", "1e999", now), http.StatusBadRequest},
		{"wrong unit", "application/json", eventBatchJSON("edge-01", "temperature", "V", "20", now), http.StatusBadRequest},
		{"explicit empty unit", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"temperature","value":20,"unit":""}]}`, http.StatusBadRequest},
		{"null unit", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"temperature","value":20,"unit":null}]}`, http.StatusBadRequest},
		{"invalid timestamp", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"yesterday","measurement_type":"temperature","value":20}]}`, http.StatusBadRequest},
		{"cross device", "application/json", eventBatchJSON("edge-02", "temperature", "C", "20", now), http.StatusBadRequest},
		{"one invalid event rejects valid peers", "application/json", `{"events":[{"event_id":"evt-01","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"temperature","value":20},{"event_id":"evt-02","device_id":"edge-01","timestamp":"2026-09-11T12:00:00Z","measurement_type":"temperature","value":126}]}`, http.StatusBadRequest},
		{"temperature outside range", "application/json", eventBatchJSON("edge-01", "temperature", "C", "125.01", now), http.StatusBadRequest},
		{"voltage outside range", "application/json", eventBatchJSON("edge-01", "voltage", "V", "48.01", now), http.StatusBadRequest},
		{"battery outside range", "application/json", eventBatchJSON("edge-01", "battery", "%", "100.01", now), http.StatusBadRequest},
		{"pressure outside range", "application/json", eventBatchJSON("edge-01", "pressure", "kPa", "1000.01", now), http.StatusBadRequest},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			called := false
			handler := validationTestHandler(validator, &called, nil)
			request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", testCase.contentType)
			request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, testCase.wantStatus)
			}
			if called {
				t.Fatal("invalid batch reached downstream admission")
			}
			wantCode := map[int]string{
				http.StatusBadRequest:            "invalid_request",
				http.StatusRequestEntityTooLarge: "payload_too_large",
				http.StatusUnsupportedMediaType:  "unsupported_media_type",
			}[testCase.wantStatus]
			if response.Header().Get("X-Request-ID") == "" || !strings.Contains(response.Body.String(), `"code":"`+wantCode+`"`) {
				t.Fatalf("error response must carry request ID and stable error code: %s", response.Body.String())
			}
		})
	}
}

func TestBatchValidatorAcceptsMaximumBatchCardinality(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	validator := NewBatchValidator(func() time.Time { return now })
	var batch ValidatedBatch
	called := false
	handler := validationTestHandler(validator, &called, &batch)
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(batchJSON(500, "edge-01", "temperature", "C", "20", now)))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted || !called || len(batch.Events) != 500 {
		t.Fatalf("500-event batch = status %d, called %t, event count %d; want accepted", response.Code, called, len(batch.Events))
	}
}

func TestBatchValidatorAcceptsExactlyOneMiBBody(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	body := eventBatchJSON("edge-01", "temperature", "C", "20", now)
	body += strings.Repeat(" ", maxTelemetryBodyBytes-len(body))
	called := false
	handler := validationTestHandler(NewBatchValidator(func() time.Time { return now }), &called, nil)
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted || !called {
		t.Fatalf("one MiB body = status %d, downstream called = %t; want accepted", response.Code, called)
	}
}

func TestBatchValidatorFailsClosedWithoutAuthentication(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	called := false
	handler := validationTestHandler(NewBatchValidator(func() time.Time { return now }), &called, nil)
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(eventBatchJSON("edge-01", "temperature", "C", "20", now)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || called || !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Fatalf("unauthenticated validation result = status %d, called %t, body %s", response.Code, called, response.Body.String())
	}
}

func TestBatchValidatorRejectsMalformedUTF8BeforeAdmission(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	body := []byte(eventBatchJSON("edge-01", "temperature", "C", "20", now))
	body = bytes.Replace(body, []byte("evt-01"), []byte{'e', 'v', 't', '-', 0xff}, 1)
	called := false
	handler := validationTestHandler(NewBatchValidator(func() time.Time { return now }), &called, nil)
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || called {
		t.Fatalf("malformed UTF-8 result = status %d, called %t; want 400 before admission", response.Code, called)
	}
}

func TestBatchValidatorAcceptsExactBoundsAndDefaultsCanonicalUnits(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	validator := NewBatchValidator(func() time.Time { return now })
	for _, testCase := range []struct {
		measurement string
		value       string
		unit        string
		wantUnit    string
	}{
		{"temperature", "-40", "", "C"},
		{"temperature", "125", "C", "C"},
		{"voltage", "0", "", "V"},
		{"voltage", "48", "V", "V"},
		{"battery", "0", "", "%"},
		{"battery", "100", "%", "%"},
		{"pressure", "0", "", "kPa"},
		{"pressure", "1000", "kPa", "kPa"},
	} {
		t.Run(testCase.measurement+"_"+testCase.value, func(t *testing.T) {
			var batch ValidatedBatch
			called := false
			handler := validationTestHandler(validator, &called, &batch)
			request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(eventBatchJSON("edge-01", testCase.measurement, testCase.unit, testCase.value, now)))
			request.Header.Set("Content-Type", "application/json; charset=utf-8")
			request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusAccepted || !called {
				t.Fatalf("valid boundary response = %d, downstream called = %t", response.Code, called)
			}
			if len(batch.Events) != 1 || batch.Events[0].Unit != testCase.wantUnit || batch.Events[0].SkipLiveAggregation {
				t.Fatalf("validated batch = %#v, want one canonical non-late event", batch)
			}
		})
	}
}

func TestBatchValidatorMarksLateEventsWithoutRejectingThem(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	validator := NewBatchValidator(func() time.Time { return now })
	var batch ValidatedBatch
	called := false
	handler := validationTestHandler(validator, &called, &batch)
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", strings.NewReader(eventBatchJSON("edge-01", "temperature", "C", "20", now.Add(-60*time.Second-time.Nanosecond))))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(withAuthorizedDevice(request.Context(), credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: "edge-01"}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted || !called || len(batch.Events) != 1 || !batch.Events[0].SkipLiveAggregation {
		t.Fatalf("late event result = status %d, called %t, batch %#v; want accepted persistence-only event", response.Code, called, batch)
	}
}

func validationTestHandler(validator BatchValidator, called *bool, batch *ValidatedBatch) http.Handler {
	return httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /test-ingestion", validator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			*called = true
			if batch != nil {
				*batch, _ = ValidatedBatchFromContext(request.Context())
			}
			w.WriteHeader(http.StatusAccepted)
		})))
	}})
}

func eventBatchJSON(deviceID, measurement, unit, value string, timestamp time.Time) string {
	unitField := ""
	if unit != "" {
		unitField = fmt.Sprintf(`,"unit":%q`, unit)
	}
	return fmt.Sprintf(`{"events":[{"event_id":"evt-01","device_id":%q,"timestamp":%q,"measurement_type":%q,"value":%s%s}]}`,
		deviceID, timestamp.Format(time.RFC3339Nano), measurement, value, unitField)
}

func batchJSON(count int, deviceID, measurement, unit, value string, timestamp time.Time) string {
	event := strings.TrimSuffix(strings.TrimPrefix(eventBatchJSON(deviceID, measurement, unit, value, timestamp), `{"events":[`), `]}`)
	return `{"events":[` + strings.TrimRight(strings.Repeat(event+",", count), ",") + `]}`
}
