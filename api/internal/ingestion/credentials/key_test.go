package credentials

import (
	"encoding/base64"
	"testing"
)

func TestParseAPIKeyAcceptsOnlyStructured32ByteSecrets(t *testing.T) {
	secret := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	credential, err := ParseAPIKey("tk_edge-01_" + secret)
	if err != nil {
		t.Fatalf("ParseAPIKey() error = %v", err)
	}
	if credential.KeyID != "edge-01" {
		t.Fatalf("KeyID = %q, want edge-01", credential.KeyID)
	}
	if credential.Secret != secret {
		t.Fatal("parsed secret does not match the supplied secret")
	}

	for _, apiKey := range []string{
		"",
		"tk__" + secret,
		"tk_key_with_underscore_" + secret,
		"tk_edge-01_short",
		"tk_edge-01_" + base64.RawURLEncoding.EncodeToString(make([]byte, 31)),
		"other_edge-01_" + secret,
	} {
		if _, err := ParseAPIKey(apiKey); err == nil {
			t.Fatalf("ParseAPIKey(%q) error = nil, want validation error", apiKey)
		}
	}
}
