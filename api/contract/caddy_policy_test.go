package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaddyEnforcesContractAccessBoundaries(t *testing.T) {
	caddyfile, err := os.ReadFile(filepath.Join("..", "..", "deploy", "Caddyfile"))
	if err != nil {
		t.Fatalf("read Caddyfile: %v", err)
	}

	config := string(caddyfile)
	if !strings.Contains(config, "header ?X-Request-ID {http.request.uuid}") {
		t.Fatal("Caddy-generated access responses must include a default X-Request-ID")
	}
	if strings.Contains(config, "handle @internal-observability") {
		t.Fatal("internal observability endpoints must not be proxied through the public Caddy gateway")
	}
	if !strings.Contains(config, "@device path /v1/telemetry/batches") {
		t.Fatal("Caddyfile must route only the ingestion endpoint as the device API-key boundary")
	}
	requireCaddyBlock(t, config, "handle @device", []string{"reverse_proxy api:8080"})
	requireCaddyBlock(t, config, "@operator", []string{"path /v1/live/* /v1/history/*"})
	requireCaddyBlock(t, config, "handle @operator", []string{
		"@operator-private",
		"remote_ip private_ranges",
		"handle @operator-private",
		"basic_auth",
		"operator {$CADDY_BASIC_AUTH_HASH}",
		"reverse_proxy api:8080",
		"respond \"Forbidden\" 403",
	})
	requireCaddyBlock(t, config, "@dashboard-private", []string{
		"not path /v1/* /healthz /readyz /metrics",
		"remote_ip private_ranges",
	})
	requireCaddyBlock(t, config, "handle @dashboard-private", []string{
		"basic_auth",
		"operator {$CADDY_BASIC_AUTH_HASH}",
		"file_server",
	})

	for _, deniedRoute := range []string{"handle @observability", "handle"} {
		requireCaddyBlock(t, config, deniedRoute, []string{"respond \"Forbidden\" 403"})
	}
}

func TestComposeKeepsTheAPIOnTheInternalNetwork(t *testing.T) {
	compose, err := os.ReadFile(filepath.Join("..", "..", "deploy", "compose.yaml"))
	if err != nil {
		t.Fatalf("read Compose configuration: %v", err)
	}

	services := string(compose)
	apiStart := strings.Index(services, "  api:\n")
	postgresStart := strings.Index(services, "\n  postgres:\n")
	if apiStart == -1 || postgresStart == -1 || postgresStart <= apiStart {
		t.Fatal("Compose configuration must define an API service before PostgreSQL")
	}

	apiService := services[apiStart:postgresStart]
	if strings.Contains(apiService, "\n    ports:") {
		t.Fatal("API port 8080 must not be published on the host")
	}
	for _, requirement := range []string{"expose: [\"8080\"]", "networks: [internal]"} {
		if !strings.Contains(apiService, requirement) {
			t.Fatalf("API service must contain %q", requirement)
		}
	}
}

func requireCaddyBlock(t *testing.T, config, name string, requirements []string) {
	t.Helper()
	start := strings.Index(config, name+" {")
	if start == -1 {
		t.Fatalf("Caddyfile must define %s", name)
	}

	block := config[start:]
	end := strings.Index(block, "\n  }")
	if end == -1 {
		t.Fatalf("Caddyfile block %s is not closed", name)
	}
	block = block[:end]
	for _, requirement := range requirements {
		if !strings.Contains(block, requirement) {
			t.Fatalf("Caddyfile block %s must contain %q", name, requirement)
		}
	}
}
