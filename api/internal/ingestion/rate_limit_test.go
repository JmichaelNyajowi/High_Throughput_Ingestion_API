package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/httpserver"
)

func TestDeviceRateLimiterChargesWholeValidatedBatchAtomically(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewDeviceRateLimiter(func() time.Time { return now })

	if result := limiter.Allow("edge-01", 98); !result.Allowed {
		t.Fatal("initial 98-event batch must be allowed")
	}
	if result := limiter.Allow("edge-01", 3); result.Allowed {
		t.Fatal("3-event batch must not be partially admitted when only two tokens remain")
	}
	if result := limiter.Allow("edge-01", 2); !result.Allowed {
		t.Fatal("rejected 3-event batch must not consume the two remaining tokens")
	}
}

func TestDeviceRateLimiterStartsFullForAnyInjectedClock(t *testing.T) {
	now := time.Time{}
	limiter := NewDeviceRateLimiter(func() time.Time { return now })
	if result := limiter.Allow("edge-01", 100); !result.Allowed {
		t.Fatal("a new device must receive the full burst even when a deterministic clock starts at zero")
	}
}

func TestDeviceRateLimiterRefillsAtTenPerSecondAndNeverExceedsBurst(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewDeviceRateLimiter(func() time.Time { return now })

	if result := limiter.Allow("edge-01", 100); !result.Allowed {
		t.Fatal("new device must begin with its 100-event burst capacity")
	}
	now = now.Add(50 * time.Millisecond)
	if result := limiter.Allow("edge-01", 1); result.Allowed || result.RetryAfter != 50*time.Millisecond {
		t.Fatalf("after 50 ms, result = %#v; want denied with 50 ms retry", result)
	}
	now = now.Add(50 * time.Millisecond)
	if result := limiter.Allow("edge-01", 1); !result.Allowed {
		t.Fatal("one token must refill after 100 ms")
	}
	now = now.Add(30 * time.Second)
	if result := limiter.Allow("edge-01", 100); !result.Allowed {
		t.Fatal("idle refill must cap at the 100-event burst capacity")
	}
}

func TestDeviceRateLimiterDoesNotRefillWhenClockMovesBackward(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewDeviceRateLimiter(func() time.Time { return now })
	if result := limiter.Allow("edge-01", 100); !result.Allowed {
		t.Fatal("consume initial burst")
	}
	now = now.Add(-time.Hour)
	if result := limiter.Allow("edge-01", 1); result.Allowed {
		t.Fatal("a backward clock must not create refill tokens")
	}
	now = now.Add(time.Hour + 100*time.Millisecond)
	if result := limiter.Allow("edge-01", 1); !result.Allowed {
		t.Fatal("the original monotonic timeline must refill one token after 100 ms")
	}
}

func TestDeviceRateLimiterIsConcurrentAndPerKey(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewDeviceRateLimiter(func() time.Time { return now })
	var accepted atomic.Int32
	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() {
			defer group.Done()
			if limiter.Allow("edge-01", 10).Allowed {
				accepted.Add(1)
			}
		}()
	}
	group.Wait()
	if accepted.Load() != 10 {
		t.Fatalf("concurrent accepted batches = %d, want exactly 10 from 100 tokens", accepted.Load())
	}
	if result := limiter.Allow("edge-02", 100); !result.Allowed {
		t.Fatal("a separate authenticated key must retain an independent burst capacity")
	}
}

func TestDeviceRateLimiterReturnsCalculableRetryAndMiddlewareRejectsBeforeAdmission(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewDeviceRateLimiter(func() time.Time { return now })
	if result := limiter.Allow("edge-01", 100); !result.Allowed {
		t.Fatal("consume initial burst")
	}
	if result := limiter.Allow("edge-01", 101); result.Allowed || result.RetryAfter != 0 {
		t.Fatalf("cost above burst capacity result = %#v; want denied with no calculable retry", result)
	}

	called := false
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /test-ingestion", limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusAccepted)
		})))
	}})
	request := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
	request = request.WithContext(withValidatedBatch(request.Context(), testValidatedBatch("edge-01", 10)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests || called {
		t.Fatalf("rate-limited result = status %d, downstream called %t; want 429 before admission", response.Code, called)
	}
	if response.Header().Get("Retry-After") != "1" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("429 headers = %#v, want Retry-After 1 and X-Request-ID", response.Header())
	}
}

func TestDeviceRateLimiterMiddlewareFailsClosedWithoutValidatedBatch(t *testing.T) {
	limiter := NewDeviceRateLimiter(time.Now)
	called := false
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /test-ingestion", limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusAccepted)
		})))
	}})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/test-ingestion", nil).WithContext(context.Background()))

	if response.Code != http.StatusBadRequest || called {
		t.Fatalf("missing validated batch result = status %d, downstream called %t; want 400 before admission", response.Code, called)
	}
}

func testValidatedBatch(keyID string, eventCount int) ValidatedBatch {
	return ValidatedBatch{
		Device: credentials.AuthorizedDevice{InternalID: "internal-01", ExternalID: "edge-01", KeyID: keyID},
		Events: make([]ValidatedEvent, eventCount),
	}
}
