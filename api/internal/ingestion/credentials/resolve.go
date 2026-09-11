package credentials

import (
	"context"
	"crypto/hmac"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthorizedDevice struct {
	InternalID string
	ExternalID string
	KeyID      string
}

func ResolveSeededCredential(ctx context.Context, pool *pgxpool.Pool, pepper, value string) (AuthorizedDevice, error) {
	apiKey, err := ParseAPIKey(value)
	if err != nil {
		return AuthorizedDevice{}, ErrUnauthorized
	}
	if len(pepper) < minimumPepperLength || pool == nil {
		return AuthorizedDevice{}, errors.New("credential store is unavailable")
	}

	var (
		device  AuthorizedDevice
		hash    []byte
		enabled bool
		active  bool
	)
	err = pool.QueryRow(ctx, `
		SELECT d.id, d.external_id, c.key_id, c.api_key_hash, c.enabled,
		       d.lifecycle_state = 'active'
		FROM api_clients c
		JOIN devices d ON d.id = c.device_id
		WHERE c.key_id = $1`, apiKey.KeyID).Scan(
		&device.InternalID,
		&device.ExternalID,
		&device.KeyID,
		&hash,
		&enabled,
		&active,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthorizedDevice{}, ErrUnauthorized
	}
	if err != nil {
		return AuthorizedDevice{}, errors.New("credential store is unavailable")
	}
	if !hmac.Equal(hash, hmacSecret(pepper, apiKey.secretBytes)) || !enabled || !active {
		return AuthorizedDevice{}, ErrUnauthorized
	}
	return device, nil
}
