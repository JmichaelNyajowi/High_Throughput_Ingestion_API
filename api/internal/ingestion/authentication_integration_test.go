package ingestion

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/example/telemetry/api/internal/platform/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestDeviceAuthenticatorUsesRealCredentialBinding(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17.3-alpine",
		postgres.WithDatabase("telemetry"),
		postgres.WithUsername("telemetry"),
		postgres.WithPassword("test-password"),
		postgres.WithInitScripts(filepath.Join(repositoryRoot(t), "migrations", "000001_init.up.sql")),
	)
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}
	testcontainers.CleanupContainer(t, container)

	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("PostgreSQL connection string: %v", err)
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	t.Cleanup(pool.Close)
	waitForPostgreSQL(t, ctx, pool)

	const externalID = "edge-01"
	pepper := strings.Repeat("p", 32)
	apiKey := integrationTestAPIKey("edge-01", 1)
	if _, err := credentials.Bootstrap(ctx, pool, pepper, []config.DeviceSeed{{
		DeviceID: externalID,
		APIKey:   apiKey,
	}}); err != nil {
		t.Fatalf("seed device credential: %v", err)
	}

	var downstreamCalls atomic.Int32
	var received credentials.AuthorizedDevice
	handler := httpserver.NewHandler(httpserver.Options{RegisterRoutes: func(mux *http.ServeMux) {
		mux.Handle("POST /test-ingestion", NewDeviceAuthenticator(pool, pepper).Middleware(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			device, found := AuthorizedDeviceFromContext(request.Context())
			if !found {
				t.Fatal("authenticated request is missing downstream device identity")
			}
			received = device
			downstreamCalls.Add(1)
			w.WriteHeader(http.StatusAccepted)
		})))
	}})

	validRequest := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
	validRequest.Header.Set("X-API-Key", apiKey)
	validResponse := httptest.NewRecorder()
	handler.ServeHTTP(validResponse, validRequest)
	if validResponse.Code != http.StatusAccepted {
		t.Fatalf("valid credential status = %d, want %d", validResponse.Code, http.StatusAccepted)
	}
	if downstreamCalls.Load() != 1 || received.InternalID == "" || received.ExternalID != externalID || received.KeyID != "edge-01" {
		t.Fatalf("downstream device identity = %#v after %d calls, want active edge-01 identity", received, downstreamCalls.Load())
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
	unknownRequest.Header.Set("X-API-Key", integrationTestAPIKey("unknown", 2))
	unknownResponse := httptest.NewRecorder()
	handler.ServeHTTP(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unknown credential status = %d, want %d", unknownResponse.Code, http.StatusUnauthorized)
	}
	if downstreamCalls.Load() != 1 {
		t.Fatalf("unknown credential reached downstream %d times, want no additional calls", downstreamCalls.Load())
	}

	if _, err := pool.Exec(ctx, `UPDATE api_clients SET enabled = false WHERE key_id = $1`, "edge-01"); err != nil {
		t.Fatalf("disable seeded credential: %v", err)
	}
	disabledRequest := httptest.NewRequest(http.MethodPost, "/test-ingestion", nil)
	disabledRequest.Header.Set("X-API-Key", apiKey)
	disabledResponse := httptest.NewRecorder()
	handler.ServeHTTP(disabledResponse, disabledRequest)
	if disabledResponse.Code != http.StatusUnauthorized {
		t.Fatalf("disabled credential status = %d, want %d", disabledResponse.Code, http.StatusUnauthorized)
	}
	if downstreamCalls.Load() != 1 {
		t.Fatalf("disabled credential reached downstream %d times, want no additional calls", downstreamCalls.Load())
	}
	if strings.Contains(disabledResponse.Body.String(), "disabled") || strings.Contains(disabledResponse.Body.String(), apiKey) {
		t.Fatalf("disabled credential response leaks authorization detail: %s", disabledResponse.Body.String())
	}
}

func waitForPostgreSQL(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := pool.Ping(ctx); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("PostgreSQL did not become ready within 30 seconds")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join("..", "..", ".."))
}

func integrationTestAPIKey(keyID string, byteValue byte) string {
	secret := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string([]byte{byteValue}), 32)))
	return "tk_" + keyID + "_" + secret
}
