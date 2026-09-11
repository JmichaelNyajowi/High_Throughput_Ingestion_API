package credentials

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestResolveSeededCredentialReturnsGenericUnauthorizedForMissingOrMalformedKeys(t *testing.T) {
	for _, apiKey := range []string{"", "not-a-structured-key"} {
		_, err := ResolveSeededCredential(context.Background(), nil, strings.Repeat("p", 32), apiKey)
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("ResolveSeededCredential(%q) error = %v, want ErrUnauthorized", apiKey, err)
		}
		if err.Error() != ErrUnauthorized.Error() {
			t.Fatalf("ResolveSeededCredential(%q) error = %q, want generic %q", apiKey, err, ErrUnauthorized)
		}
	}
}
