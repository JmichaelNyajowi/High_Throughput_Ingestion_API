package ingestion

import (
	"context"
	"errors"
	"net/http"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authorizedDeviceContextKey struct{}

// DeviceAuthenticator admits requests to the ingestion route only after their
// seeded device credential has resolved to one enabled, active device.
type DeviceAuthenticator struct {
	pool   *pgxpool.Pool
	pepper string
}

func NewDeviceAuthenticator(pool *pgxpool.Pool, pepper string) DeviceAuthenticator {
	return DeviceAuthenticator{pool: pool, pepper: pepper}
}

func (authenticator DeviceAuthenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.Header.Values("X-API-Key")
		if len(values) != 1 {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication failed")
			return
		}

		device, err := credentials.ResolveSeededCredential(r.Context(), authenticator.pool, authenticator.pepper, values[0])
		if errors.Is(err, credentials.ErrUnauthorized) {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication failed")
			return
		}
		if err != nil {
			httpserver.WriteError(w, r, http.StatusServiceUnavailable, "unavailable", "Service temporarily unavailable")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authorizedDeviceContextKey{}, device)))
	})
}

// AuthorizedDeviceFromContext returns the device identity established by the
// device-authentication boundary. It never reads an untrusted request value.
func AuthorizedDeviceFromContext(ctx context.Context) (credentials.AuthorizedDevice, bool) {
	device, found := ctx.Value(authorizedDeviceContextKey{}).(credentials.AuthorizedDevice)
	return device, found
}
