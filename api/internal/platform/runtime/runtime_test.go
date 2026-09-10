package runtime

import (
	"context"
	"testing"

	"github.com/example/telemetry/api/internal/platform/httpserver"
)

func TestShutdownPausesReadinessBeforeServerAndDependencyTeardown(t *testing.T) {
	readiness := httpserver.NewReadiness()
	order := make([]string, 0, 2)
	server := shutdownFunc(func(context.Context) error {
		if readiness.Accepting() {
			t.Fatal("readiness must pause before HTTP shutdown starts")
		}
		order = append(order, "server")
		return nil
	})
	drainer := shutdownFunc(func(context.Context) error {
		order = append(order, "drainer")
		return nil
	})

	service := New(server, readiness, drainer)
	if err := service.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if got, want := len(order), 2; got != want || order[0] != "server" || order[1] != "drainer" {
		t.Fatalf("shutdown order = %v, want [server drainer]", order)
	}
}

type shutdownFunc func(context.Context) error

func (fn shutdownFunc) Shutdown(ctx context.Context) error {
	return fn(ctx)
}
