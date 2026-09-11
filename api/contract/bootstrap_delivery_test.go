package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapDocumentationUsesExecutableCommands(t *testing.T) {
	setup := readProjectFile(t, "..", "..", "docs", "setup.md")
	if !strings.Contains(setup, "`go -C api run ./cmd/seed`") {
		t.Fatal("setup must invoke the seed command from the api module")
	}
	if strings.Contains(setup, "`go run ./api/cmd/seed`") {
		t.Fatal("setup must not invoke the nested Go module from the repository root")
	}
	if !strings.Contains(setup, "`go -C api run ./cmd/api`") || strings.Contains(setup, "`go run ./api/cmd/api`") {
		t.Fatal("setup must invoke the API command from the api module")
	}
	if !strings.Contains(setup, "`docker compose -f deploy/compose.yaml --env-file .env run --rm --entrypoint /seed api`") {
		t.Fatal("Compose seed instruction must override the API image entrypoint")
	}

	dockerfile := readProjectFile(t, "..", "Dockerfile")
	if !strings.Contains(dockerfile, "ENTRYPOINT [\"/api\"]") || !strings.Contains(dockerfile, "COPY --from=build /out/seed /seed") {
		t.Fatal("seed delivery test requires the API image to contain /seed behind the /api entrypoint")
	}
}

func readProjectFile(t *testing.T, path ...string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(path...))
	if err != nil {
		t.Fatalf("read %s: %v", filepath.Join(path...), err)
	}
	return string(contents)
}
