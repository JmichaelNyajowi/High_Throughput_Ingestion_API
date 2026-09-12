package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestHealthEndpointsConformToPublishedContract(t *testing.T) {
	handler := NewHandler(Options{})
	for _, endpoint := range []struct {
		path string
		body map[string]string
	}{
		{"/healthz", map[string]string{"status": "ok"}},
		{"/readyz", map[string]string{"status": "ready", "admission": "accepting", "postgres": "unknown", "redis": "unknown"}},
	} {
		t.Run(endpoint.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, endpoint.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", contentType)
			}
			if requestID := response.Header().Get("X-Request-ID"); !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(requestID) {
				t.Fatalf("X-Request-ID = %q, want UUIDv4", requestID)
			}

			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode JSON body: %v", err)
			}
			if len(body) != len(endpoint.body) {
				t.Fatalf("body = %#v, want %#v", body, endpoint.body)
			}
			for key, want := range endpoint.body {
				if got := body[key]; got != want {
					t.Fatalf("body[%q] = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestHandlerPropagatesRequestIDAndLogsOnlySafeRequestFields(t *testing.T) {
	var logs bytes.Buffer
	requestID := "796eaf08-135a-4e6d-932a-171b8f3853d3"
	handler := NewHandler(Options{Logger: slog.New(slog.NewJSONHandler(&logs, nil))})
	request := httptest.NewRequest(http.MethodGet, "/healthz?credential=must-not-log", strings.NewReader("raw-payload-must-not-log"))
	request.Header.Set("X-Request-ID", requestID)
	request.Header.Set("X-API-Key", "secret-device-key-must-not-log")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got := response.Header().Get("X-Request-ID"); got != requestID {
		t.Fatalf("X-Request-ID = %q, want propagated ID %q", got, requestID)
	}
	output := logs.String()
	if !strings.Contains(output, requestID) {
		t.Fatalf("request log does not contain propagated request ID: %s", output)
	}
	for _, forbidden := range []string{"secret-device-key-must-not-log", "raw-payload-must-not-log", "credential=must-not-log"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("request log must not contain %q: %s", forbidden, output)
		}
	}
}

func TestHandlerRecoversPanicsWithSafeJSONError(t *testing.T) {
	handler := NewHandler(Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.HandleFunc("GET /panic", func(http.ResponseWriter, *http.Request) {
			panic("secret panic detail")
		})
	}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if strings.Contains(response.Body.String(), "secret panic detail") {
		t.Fatalf("recovery body leaks panic detail: %s", response.Body.String())
	}
	var body struct {
		RequestID string `json:"request_id"`
		Error     struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode recovery response: %v", err)
	}
	if body.RequestID == "" || body.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("response must return its request ID in the header and envelope: %#v", body)
	}
	if body.Error.Code != "unavailable" || body.Error.Message != "Service temporarily unavailable" {
		t.Fatalf("recovery error = %#v, want safe unavailable error", body.Error)
	}
}

func TestReadinessPausesBeforeShutdownWhileLivenessRemainsAvailable(t *testing.T) {
	readiness := NewReadiness()
	handler := NewHandler(Options{Readiness: readiness})
	readiness.Pause()

	readyResponse := httptest.NewRecorder()
	handler.ServeHTTP(readyResponse, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if readyResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status = %d, want %d", readyResponse.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(readyResponse.Body.String(), `"status":"unavailable"`) || !strings.Contains(readyResponse.Body.String(), `"admission":"paused"`) {
		t.Fatalf("readyz body = %s, want unavailable paused state", readyResponse.Body.String())
	}

	livenessResponse := httptest.NewRecorder()
	handler.ServeHTTP(livenessResponse, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if livenessResponse.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", livenessResponse.Code, http.StatusOK)
	}
}

func TestRouterErrorsStillIncludeGeneratedRequestIDs(t *testing.T) {
	handler := NewHandler(Options{})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/healthz", nil))

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if requestID := response.Header().Get("X-Request-ID"); !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(requestID) {
		t.Fatalf("X-Request-ID = %q, want generated UUIDv4", requestID)
	}
}
