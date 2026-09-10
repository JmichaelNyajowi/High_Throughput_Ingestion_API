package config

import (
	"testing"
	"time"
)

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

func TestLoadProvidesExplicitRuntimeAndDependencyTimeouts(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	for _, name := range []string{
		"API_READ_HEADER_TIMEOUT",
		"API_READ_TIMEOUT",
		"API_WRITE_TIMEOUT",
		"API_IDLE_TIMEOUT",
		"API_POSTGRES_TIMEOUT",
		"API_REDIS_TIMEOUT",
		"API_SHUTDOWN_TIMEOUT",
	} {
		t.Setenv(name, "")
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Server.ReadHeaderTimeout <= 0 || config.Server.ReadTimeout <= 0 || config.Server.WriteTimeout <= 0 || config.Server.IdleTimeout <= 0 {
		t.Fatalf("server timeouts must all be positive: %#v", config.Server)
	}
	if config.Dependencies.PostgresTimeout <= 0 || config.Dependencies.RedisTimeout <= 0 || config.ShutdownTimeout <= 0 {
		t.Fatalf("dependency and shutdown timeouts must all be positive: %#v", config)
	}
}

func TestLoadRejectsInvalidRuntimeTimeout(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("API_READ_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid timeout error")
	}
}

func TestLoadUsesValidatedTimeoutOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("API_READ_TIMEOUT", "3s")
	t.Setenv("API_POSTGRES_TIMEOUT", "750ms")
	t.Setenv("API_SHUTDOWN_TIMEOUT", "12s")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Server.ReadTimeout != 3*time.Second || config.Dependencies.PostgresTimeout != 750*time.Millisecond || config.ShutdownTimeout != 12*time.Second {
		t.Fatalf("timeout overrides were not applied: %#v", config)
	}
}
