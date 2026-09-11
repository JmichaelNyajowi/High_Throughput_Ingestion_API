package credentials

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const minimumPepperLength = 32

type preparedSeed struct {
	deviceID   string
	name       string
	deviceType string
	keyID      string
	keyHash    []byte
}

func Bootstrap(ctx context.Context, pool *pgxpool.Pool, pepper string, seeds []config.DeviceSeed) (int, error) {
	prepared, err := prepareSeeds(pepper, seeds)
	if err != nil {
		return 0, err
	}
	if len(prepared) == 0 {
		return 0, nil
	}
	if pool == nil {
		return 0, errors.New("credential bootstrap database is unavailable")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, errors.New("begin credential bootstrap")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, seed := range prepared {
		var internalDeviceID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO devices (external_id, name, device_type, lifecycle_state)
			VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), 'active')
			ON CONFLICT (external_id) DO UPDATE
			SET name = EXCLUDED.name,
				device_type = EXCLUDED.device_type,
				lifecycle_state = 'active',
				updated_at = now()
			RETURNING id`, seed.deviceID, seed.name, seed.deviceType).Scan(&internalDeviceID); err != nil {
			return 0, errors.New("store seeded device")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO api_clients (key_id, api_key_hash, device_id, enabled)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (device_id) DO UPDATE
			SET key_id = EXCLUDED.key_id,
				api_key_hash = EXCLUDED.api_key_hash,
				enabled = true`, seed.keyID, seed.keyHash, internalDeviceID); err != nil {
			return 0, errors.New("store seeded credential")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, errors.New("commit credential bootstrap")
	}
	return len(prepared), nil
}

func prepareSeeds(pepper string, seeds []config.DeviceSeed) ([]preparedSeed, error) {
	if len(seeds) == 0 {
		return nil, nil
	}
	if len(pepper) < minimumPepperLength {
		return nil, errors.New("credential bootstrap pepper must be at least 32 bytes")
	}

	prepared := make([]preparedSeed, 0, len(seeds))
	deviceIDs := make(map[string]struct{}, len(seeds))
	keyIDs := make(map[string]struct{}, len(seeds))
	for _, seed := range seeds {
		if err := validateDeviceSeed(seed); err != nil {
			return nil, err
		}
		apiKey, err := ParseAPIKey(seed.APIKey)
		if err != nil {
			return nil, errors.New("credential bootstrap seed has an invalid API key")
		}
		if _, exists := deviceIDs[seed.DeviceID]; exists {
			return nil, errors.New("credential bootstrap has duplicate device identifiers")
		}
		if _, exists := keyIDs[apiKey.KeyID]; exists {
			return nil, errors.New("credential bootstrap has duplicate key identifiers")
		}
		deviceIDs[seed.DeviceID] = struct{}{}
		keyIDs[apiKey.KeyID] = struct{}{}
		prepared = append(prepared, preparedSeed{
			deviceID:   seed.DeviceID,
			name:       seed.Name,
			deviceType: seed.DeviceType,
			keyID:      apiKey.KeyID,
			keyHash:    hmacSecret(pepper, apiKey.secretBytes),
		})
	}
	return prepared, nil
}

func validateDeviceSeed(seed config.DeviceSeed) error {
	if seed.DeviceID == "" || len(seed.DeviceID) > 128 || strings.TrimSpace(seed.DeviceID) != seed.DeviceID {
		return errors.New("credential bootstrap seed has an invalid device identifier")
	}
	if len(seed.Name) > 128 || len(seed.DeviceType) > 64 {
		return errors.New("credential bootstrap seed has an invalid device description")
	}
	return nil
}

func hmacSecret(pepper string, secret []byte) []byte {
	hash := hmac.New(sha256.New, []byte(pepper))
	_, _ = hash.Write(secret)
	return hash.Sum(nil)
}
