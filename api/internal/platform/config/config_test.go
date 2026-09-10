package config

import "testing"

func TestLoadRequiresSecretsInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("POSTGRES_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("TELEMETRY_API_KEY_PEPPER", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want required production configuration error")
	}
}

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("API_LISTEN_ADDR", "")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.ListenAddr != ":8080" {
		t.Fatalf("ListenAddr = %q, want :8080", config.ListenAddr)
	}
}
