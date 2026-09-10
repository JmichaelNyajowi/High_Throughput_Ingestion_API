package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpointsReturnNoContent(t *testing.T) {
	handler := NewHandler()
	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
			}
		})
	}
}
