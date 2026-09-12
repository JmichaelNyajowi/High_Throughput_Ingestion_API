package ingestion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/telemetry/api/internal/platform/httpserver"
)

func TestTelemetryAdmissionEndpointAcknowledgesOnlyEnqueuedValidatedBatch(t *testing.T) {
	queue := &recordingAdmissionQueue{}
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /v1/telemetry/batches", NewTelemetryAdmissionEndpoint(queue))
	}})
	request := httptest.NewRequest(http.MethodPost, "/v1/telemetry/batches", nil)
	request = request.WithContext(withValidatedBatch(request.Context(), testValidatedBatch("edge-01", 2)))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("response headers = %#v, want JSON and request ID", response.Header())
	}
	var body admissionAccepted
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 202 response: %v", err)
	}
	if body.RequestID != response.Header().Get("X-Request-ID") || body.AcceptedEvents != 2 || body.Status != "accepted" {
		t.Fatalf("202 body = %#v, want request ID, two accepted events, and accepted status", body)
	}
	if len(queue.batches) != 1 || len(queue.batches[0].Events) != 2 {
		t.Fatalf("response/body queue state = %q %#v, want one admitted 2-event batch", response.Body.String(), queue.batches)
	}
}

func TestTelemetryAdmissionEndpointReturnsStable503WithoutPartialAdmission(t *testing.T) {
	queue := &unavailableAdmissionQueue{}
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /v1/telemetry/batches", NewTelemetryAdmissionEndpoint(queue))
	}})
	request := httptest.NewRequest(http.MethodPost, "/v1/telemetry/batches", nil)
	request = request.WithContext(withValidatedBatch(request.Context(), testValidatedBatch("edge-01", 2)))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable || queue.calls != 1 || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("503 result = status %d calls %d headers %#v", response.Code, queue.calls, response.Header())
	}
	var body struct {
		RequestID string `json:"request_id"`
		Error     struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 503 response: %v", err)
	}
	if body.RequestID != response.Header().Get("X-Request-ID") || body.Error.Code != "unavailable" || body.Error.Message != "Service temporarily unavailable" {
		t.Fatalf("503 body = %#v, want typed unavailable response with request ID", body)
	}
}

func TestTelemetryAdmissionEndpointRejectsBackpressuredBatchBeforeQueue(t *testing.T) {
	queue := &recordingAdmissionQueue{}
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /v1/telemetry/batches", NewTelemetryAdmissionEndpoint(queue, blockedAdmissionGuard{}))
	}})
	request := httptest.NewRequest(http.MethodPost, "/v1/telemetry/batches", nil)
	request = request.WithContext(withValidatedBatch(request.Context(), testValidatedBatch("edge-01", 2)))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable || len(queue.batches) != 0 || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("backpressure result = status %d queue %#v headers %#v", response.Code, queue.batches, response.Header())
	}
}

func TestTelemetryAdmissionRouteRejectsUnauthenticatedRequestBeforeQueue(t *testing.T) {
	queue := &recordingAdmissionQueue{}
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /v1/telemetry/batches", NewTelemetryAdmissionRoute(NewDeviceAuthenticator(nil, "test-pepper"), NewBatchValidator(nil), NewDeviceRateLimiter(nil), queue))
	}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/telemetry/batches", nil))

	if response.Code != http.StatusUnauthorized || len(queue.batches) != 0 || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("unauthenticated route result = status %d queue %#v headers %#v", response.Code, queue.batches, response.Header())
	}
}

type recordingAdmissionQueue struct {
	batches []ValidatedBatch
}

func (queue *recordingAdmissionQueue) Admit(batch ValidatedBatch) error {
	queue.batches = append(queue.batches, batch)
	return nil
}

type unavailableAdmissionQueue struct {
	calls int
}

func (queue *unavailableAdmissionQueue) Admit(ValidatedBatch) error {
	queue.calls++
	return unavailableAdmissionError{}
}

type unavailableAdmissionError struct{}

func (unavailableAdmissionError) Error() string            { return "queue full" }
func (unavailableAdmissionError) AdmissionStatusCode() int { return http.StatusServiceUnavailable }
func (unavailableAdmissionError) AdmissionCode() string    { return "unavailable" }

type blockedAdmissionGuard struct{}

func (blockedAdmissionGuard) AllowsAdmission() bool { return false }
