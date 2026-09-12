package ingestion

import (
	"net/http"

	"github.com/example/telemetry/api/internal/platform/httpserver"
)

// BatchAdmissionQueue is the Ingestion-to-Telemetry seam. Its implementation
// owns in-memory sharding and processing; admission only sends a fully
// authenticated, validated, and rate-limited batch to it.
type BatchAdmissionQueue interface {
	Admit(ValidatedBatch) error
}

// AdmissionGuard is a cross-module safety seam. Telemetry owns the
// backpressure state; Ingestion checks it after validation/rate limiting and
// before it writes to the bounded queue.
type AdmissionGuard interface {
	AllowsAdmission() bool
}

// admissionFailure permits the queue to return the published typed overload
// response without exposing its package implementation to Ingestion.
type admissionFailure interface {
	error
	AdmissionStatusCode() int
	AdmissionCode() string
}

type telemetryAdmissionEndpoint struct {
	queue BatchAdmissionQueue
	guard AdmissionGuard
}

// NewTelemetryAdmissionEndpoint returns the final admission handler. It must
// be reached only after authentication, validation, and rate limiting.
func NewTelemetryAdmissionEndpoint(queue BatchAdmissionQueue, guards ...AdmissionGuard) http.Handler {
	return telemetryAdmissionEndpoint{queue: queue, guard: firstAdmissionGuard(guards)}
}

// NewTelemetryAdmissionRoute composes the mandatory device-admission order:
// authentication, whole-batch validation, event-cost rate limiting, then queue
// admission. It performs no Redis or PostgreSQL operation on the request path.
func NewTelemetryAdmissionRoute(authenticator DeviceAuthenticator, validator BatchValidator, limiter *DeviceRateLimiter, queue BatchAdmissionQueue, guards ...AdmissionGuard) http.Handler {
	return authenticator.Middleware(validator.Middleware(limiter.Middleware(NewTelemetryAdmissionEndpoint(queue, guards...))))
}

func (endpoint telemetryAdmissionEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	batch, found := ValidatedBatchFromContext(r.Context())
	if !found || endpoint.queue == nil {
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Service temporarily unavailable")
		return
	}
	if endpoint.guard != nil && !endpoint.guard.AllowsAdmission() {
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Service temporarily unavailable")
		return
	}
	batch.RequestID = httpserver.RequestID(r.Context())
	if err := endpoint.queue.Admit(batch); err != nil {
		if failure, ok := err.(admissionFailure); ok {
			httpserver.WriteError(w, r, failure.AdmissionStatusCode(), failure.AdmissionCode(), "Service temporarily unavailable")
			return
		}
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Service temporarily unavailable")
		return
	}

	httpserver.WriteJSON(w, http.StatusAccepted, admissionAccepted{
		RequestID:      httpserver.RequestID(r.Context()),
		AcceptedEvents: len(batch.Events),
		Status:         "accepted",
	})
}

func firstAdmissionGuard(guards []AdmissionGuard) AdmissionGuard {
	if len(guards) == 0 {
		return nil
	}
	return guards[0]
}

type admissionAccepted struct {
	RequestID      string `json:"request_id"`
	AcceptedEvents int    `json:"accepted_events"`
	Status         string `json:"status"`
}
