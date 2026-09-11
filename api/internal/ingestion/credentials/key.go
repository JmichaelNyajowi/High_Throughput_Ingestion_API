package credentials

import (
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
)

const (
	keyPrefix        = "tk"
	secretByteLength = 32
)

var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

type APIKey struct {
	KeyID       string
	Secret      string
	secretBytes []byte
}

func ParseAPIKey(value string) (APIKey, error) {
	parts := strings.SplitN(value, "_", 3)
	if len(parts) != 3 || parts[0] != keyPrefix || !keyIDPattern.MatchString(parts[1]) {
		return APIKey{}, errors.New("invalid API key format")
	}

	secret, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(secret) != secretByteLength {
		return APIKey{}, errors.New("invalid API key format")
	}

	return APIKey{KeyID: parts[1], Secret: parts[2], secretBytes: secret}, nil
}
