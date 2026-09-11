package credentials

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestBootstrapPersistsOnlyHMACCredentialRecordsAgainstRealSchema(t *testing.T) {
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
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	firstKey := testAPIKey("edge-01", 1)
	seeded, err := Bootstrap(ctx, pool, strings.Repeat("p", 32), []config.DeviceSeed{{
		DeviceID:   "edge-01",
		Name:       "Boiler inlet",
		DeviceType: "gateway",
		APIKey:     firstKey,
	}})
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if seeded != 1 {
		t.Fatalf("Bootstrap() seeded = %d, want 1", seeded)
	}

	parsedFirstKey, err := ParseAPIKey(firstKey)
	if err != nil {
		t.Fatalf("ParseAPIKey() error = %v", err)
	}
	var (
		externalID string
		lifecycle  string
		keyID      string
		hash       []byte
		enabled    bool
	)
	err = pool.QueryRow(ctx, `
		SELECT d.external_id, d.lifecycle_state::text, c.key_id, c.api_key_hash, c.enabled
		FROM devices d
		JOIN api_clients c ON c.device_id = d.id
		WHERE d.external_id = $1`, "edge-01").Scan(&externalID, &lifecycle, &keyID, &hash, &enabled)
	if err != nil {
		t.Fatalf("read seeded credential: %v", err)
	}
	if externalID != "edge-01" || lifecycle != "active" || keyID != parsedFirstKey.KeyID || !enabled {
		t.Fatalf("seeded credential state = (%q, %q, %q, %t), want active enabled edge-01 binding", externalID, lifecycle, keyID, enabled)
	}
	if len(hash) != 32 || string(hash) == parsedFirstKey.Secret || strings.Contains(string(hash), parsedFirstKey.Secret) {
		t.Fatal("database must retain only a 32-byte HMAC hash, never the credential secret")
	}
	if !hmac.Equal(hash, hmacSecret(strings.Repeat("p", 32), parsedFirstKey.secretBytes)) {
		t.Fatal("database credential hash does not match the seeded HMAC-SHA-256 value")
	}
	device, err := ResolveSeededCredential(ctx, pool, strings.Repeat("p", 32), firstKey)
	if err != nil {
		t.Fatalf("ResolveSeededCredential() error = %v", err)
	}
	if device.ExternalID != "edge-01" || device.KeyID != parsedFirstKey.KeyID || device.InternalID == "" {
		t.Fatalf("resolved credential = %#v, want the active edge-01 binding", device)
	}
	if _, err := ResolveSeededCredential(ctx, pool, strings.Repeat("p", 32), testAPIKey("unknown", 3)); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("unknown credential error = %v, want ErrUnauthorized", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE api_clients SET enabled = false WHERE key_id = $1`, parsedFirstKey.KeyID); err != nil {
		t.Fatalf("disable credential: %v", err)
	}
	if _, err := ResolveSeededCredential(ctx, pool, strings.Repeat("p", 32), firstKey); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("disabled credential error = %v, want ErrUnauthorized", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE api_clients SET enabled = true WHERE key_id = $1`, parsedFirstKey.KeyID); err != nil {
		t.Fatalf("enable credential: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE devices SET lifecycle_state = 'disabled' WHERE external_id = $1`, "edge-01"); err != nil {
		t.Fatalf("disable device: %v", err)
	}
	if _, err := ResolveSeededCredential(ctx, pool, strings.Repeat("p", 32), firstKey); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("inactive-device credential error = %v, want ErrUnauthorized", err)
	}

	secondKey := testAPIKey("edge-02", 2)
	if _, err := Bootstrap(ctx, pool, strings.Repeat("p", 32), []config.DeviceSeed{{DeviceID: "edge-01", APIKey: secondKey}}); err != nil {
		t.Fatalf("Bootstrap() replacement error = %v", err)
	}
	parsedSecondKey, err := ParseAPIKey(secondKey)
	if err != nil {
		t.Fatalf("ParseAPIKey() replacement key error = %v", err)
	}
	var credentialCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM api_clients c JOIN devices d ON d.id = c.device_id WHERE d.external_id = $1`, "edge-01").Scan(&credentialCount)
	if err != nil {
		t.Fatalf("count replaced credential: %v", err)
	}
	if credentialCount != 1 {
		t.Fatalf("credential count = %d, want exactly one", credentialCount)
	}
	var replacementKeyID string
	if err := pool.QueryRow(ctx, `SELECT c.key_id FROM api_clients c JOIN devices d ON d.id = c.device_id WHERE d.external_id = $1`, "edge-01").Scan(&replacementKeyID); err != nil {
		t.Fatalf("read replacement credential: %v", err)
	}
	if replacementKeyID != parsedSecondKey.KeyID {
		t.Fatalf("replacement key ID = %q, want %q", replacementKeyID, parsedSecondKey.KeyID)
	}
}

func TestBootstrapRejectsInvalidSeedBeforeWriting(t *testing.T) {
	const plaintext = "plaintext-secret"
	_, err := Bootstrap(context.Background(), nil, strings.Repeat("p", 32), []config.DeviceSeed{{
		DeviceID: "edge-01",
		APIKey:   plaintext,
	}})
	if err == nil {
		t.Fatal("Bootstrap() error = nil, want invalid seed error")
	}
	if strings.Contains(err.Error(), plaintext) {
		t.Fatal("bootstrap validation error must not echo a plaintext credential")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join("..", "..", "..", ".."))
}

func testAPIKey(keyID string, byteValue byte) string {
	secret := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string([]byte{byteValue}), 32)))
	return "tk_" + keyID + "_" + secret
}
