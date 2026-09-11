package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid credential bootstrap configuration")
		os.Exit(1)
	}
	if len(cfg.DeviceSeeds) == 0 {
		logger.Info("no device credential seeds configured")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Dependencies.PostgresTimeout)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		logger.Error("credential bootstrap database is unavailable")
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		logger.Error("credential bootstrap database is unavailable")
		os.Exit(1)
	}

	seeded, err := credentials.Bootstrap(ctx, pool, cfg.APIKeyPepper, cfg.DeviceSeeds)
	if err != nil {
		logger.Error("credential bootstrap failed")
		os.Exit(1)
	}
	logger.Info("device credential bootstrap completed", "seeded_devices", seeded)
}
