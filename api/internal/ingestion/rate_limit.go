package ingestion

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/example/telemetry/api/internal/platform/httpserver"
	"golang.org/x/time/rate"
)

const (
	deviceTokenRefillRate = rate.Limit(10)
	deviceTokenBurst      = 100
)

type RateLimitResult struct {
	Allowed    bool
	RetryAfter time.Duration
}

type deviceRateBucket struct {
	mu      sync.Mutex
	limiter *rate.Limiter
	lastNow time.Time
}

// DeviceRateLimiter is deliberately in-memory and instance-local. A future
// multi-instance deployment must not treat it as a distributed quota service.
type DeviceRateLimiter struct {
	now     func() time.Time
	buckets sync.Map // map[string]*deviceRateBucket, keyed by authenticated API key ID
}

func NewDeviceRateLimiter(now func() time.Time) *DeviceRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &DeviceRateLimiter{now: now}
}

// Allow atomically charges one token per validated event for one authenticated
// key. A rejected batch consumes no tokens.
func (limiter *DeviceRateLimiter) Allow(keyID string, eventCount int) RateLimitResult {
	if keyID == "" || eventCount < 1 {
		return RateLimitResult{}
	}

	now := limiter.now()
	value, found := limiter.buckets.Load(keyID)
	if !found {
		candidate := newDeviceRateBucket(now)
		value, _ = limiter.buckets.LoadOrStore(keyID, candidate)
	}
	bucket := value.(*deviceRateBucket)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	now = bucket.monotonicNow(now)
	if bucket.limiter.AllowN(now, eventCount) {
		return RateLimitResult{Allowed: true}
	}
	if eventCount > deviceTokenBurst {
		return RateLimitResult{}
	}

	reservation := bucket.limiter.ReserveN(now, eventCount)
	if !reservation.OK() {
		return RateLimitResult{}
	}
	retryAfter := reservation.DelayFrom(now)
	reservation.CancelAt(now)
	return RateLimitResult{RetryAfter: retryAfter}
}

func (limiter *DeviceRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batch, found := ValidatedBatchFromContext(r.Context())
		if !found {
			httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid telemetry batch")
			return
		}

		result := limiter.Allow(batch.Device.KeyID, len(batch.Events))
		if !result.Allowed {
			if result.RetryAfter > 0 {
				seconds := int64(result.RetryAfter / time.Second)
				if result.RetryAfter%time.Second != 0 {
					seconds++
				}
				if seconds < 1 {
					seconds = 1
				}
				w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
			}
			httpserver.WriteError(w, r, http.StatusTooManyRequests, "rate_limited", "Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r.WithContext(withValidatedBatch(r.Context(), batch)))
	})
}

func newDeviceRateBucket(now time.Time) *deviceRateBucket {
	bucket := &deviceRateBucket{limiter: rate.NewLimiter(deviceTokenRefillRate, deviceTokenBurst), lastNow: now}
	// rate.Limiter starts empty. A new device is intentionally granted the full
	// approved burst capacity immediately, including under a zero-based test
	// clock.
	_ = bucket.limiter.AllowN(now.Add(-10*time.Second), 0)
	_ = bucket.limiter.AllowN(now, 0)
	return bucket
}

func (bucket *deviceRateBucket) monotonicNow(now time.Time) time.Time {
	if now.Before(bucket.lastNow) {
		return bucket.lastNow
	}
	bucket.lastNow = now
	return now
}
