package ingestion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/telemetry/api/internal/platform/httpserver"
)

func TestDeviceAuthenticatorRejectsUnauthenticatedRequestsBeforeDownstreamAdmission(t *testing.T) {
	t.Parallel()

	for name, setKey := range map[string]func(*http.Request){
		"missing": func(*http.Request) {},
		"malformed": func(request *http.Request) {
			request.Header.Set("X-API-Key", "not-a-structured-key")
		},
		"multiple": func(request *http.Request) {
			request.Header.Add("X-API-Key", "tk_edge-01_"+strings.Repeat("A", 43))
			request.Header.Add("X-API-Key", "tk_edge-02_"+strings.Repeat("A", 43))
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			called := false
			authenticator := NewDeviceAuthenticator(nil, strings.Repeat("p", 32))
			handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
				mux.Handle("POST /test-ingestion", authenticator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					called = true
					w.WriteHeader(http.StatusAccepted)
				})))
			}})

			request := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
			setKey(request)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if called {
				t.Fatal("unauthenticated request reached downstream admission")
			}
			if response.Header().Get("X-Request-ID") == "" {
				t.Fatal("401 response must include X-Request-ID")
			}

			var body struct {
				RequestID string `json:"request_id"`
				Error     struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.RequestID == "" || body.Error.Code != "unauthorized" || body.Error.Message != "Authentication failed" {
				t.Fatalf("401 response = %#v, want a generic authentication failure", body)
			}
		})
	}
}

func TestDeviceAuthenticatorTreatsCredentialStoreFailuresAsUnavailable(t *testing.T) {
	t.Parallel()

	called := false
	authenticator := NewDeviceAuthenticator(nil, strings.Repeat("p", 32))
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /test-ingestion", authenticator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusAccepted)
		})))
	}})

	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
	request.Header.Set("X-API-Key", "tk_edge-01_"+strings.Repeat("A", 43))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if called {
		t.Fatal("request with unavailable credential store reached downstream admission")
	}
}

func TestAuthorizedDeviceFromContextIsAbsentWithoutAuthentication(t *testing.T) {
	t.Parallel()
	if _, found := AuthorizedDeviceFromContext(httptest.NewRequest(http.MethodPost, "/", nil).Context()); found {
		t.Fatal("unauthenticated context must not contain a device identity")
	}
}
