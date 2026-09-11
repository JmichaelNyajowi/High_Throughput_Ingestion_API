package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
)

const requestIDHeader = "X-Request-ID"

var requestIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type requestIDContextKey struct{}

type Options struct {
	Logger         *slog.Logger
	Readiness      *Readiness
	RegisterRoutes func(*http.ServeMux)
}

type Readiness struct {
	accepting atomic.Bool
}

func NewReadiness() *Readiness {
	readiness := &Readiness{}
	readiness.accepting.Store(true)
	return readiness
}

func (r *Readiness) Accepting() bool {
	return r.accepting.Load()
}

func (r *Readiness) Pause() {
	r.accepting.Store(false)
}

func NewHandler(options Options) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	readiness := options.Readiness
	if readiness == nil {
		readiness = NewReadiness()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthStatus{Status: "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !readiness.Accepting() {
			writeJSON(w, http.StatusServiceUnavailable, readinessStatus{Status: "unavailable", Admission: "paused"})
			return
		}
		writeJSON(w, http.StatusOK, readinessStatus{Status: "ready", Admission: "accepting"})
	})
	if options.RegisterRoutes != nil {
		options.RegisterRoutes(mux)
	}

	return withRequestID(logRequests(logger, recoverPanics(logger, mux)))
}

type healthStatus struct {
	Status string `json:"status"`
}

type readinessStatus struct {
	Status    string `json:"status"`
	Admission string `json:"admission"`
}

type errorEnvelope struct {
	RequestID string `json:"request_id"`
	Error     struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if !requestIDPattern.MatchString(requestID) {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDContextKey{}, requestID)))
	})
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() == nil {
				return
			}
			logger.Error("request panic recovered", "request_id", RequestID(r.Context()), "method", r.Method, "path", r.URL.Path)
			if response, ok := w.(interface{ Written() bool }); ok && response.Written() {
				return
			}
			writeSafeError(w, r, http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &responseRecorder{ResponseWriter: w}
		startedAt := time.Now()
		next.ServeHTTP(response, r)
		logger.Info("request completed",
			"request_id", RequestID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", response.Status(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.written {
		return
	}
	w.status = status
	w.written = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *responseRecorder) Written() bool {
	return w.written
}

func (w *responseRecorder) Status() int {
	if !w.written {
		return http.StatusOK
	}
	return w.status
}

func writeSafeError(w http.ResponseWriter, r *http.Request, status int) {
	WriteError(w, r, status, "unavailable", "Service temporarily unavailable")
}

// WriteError writes the published JSON error envelope without exposing internal
// error values. Callers must choose a contract-approved code and safe message.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	payload := errorEnvelope{RequestID: RequestID(r.Context())}
	payload.Error.Code = code
	payload.Error.Message = message
	writeJSON(w, status, payload)
}

// WriteJSON writes a contract response with the shared JSON content type.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return hex.EncodeToString(bytes[0:4]) + "-" +
		hex.EncodeToString(bytes[4:6]) + "-" +
		hex.EncodeToString(bytes[6:8]) + "-" +
		hex.EncodeToString(bytes[8:10]) + "-" +
		hex.EncodeToString(bytes[10:16])
}
