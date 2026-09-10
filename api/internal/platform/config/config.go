package config

import (
	"fmt"
	"os"
)

type Config struct {
	ListenAddr   string
	Environment  string
	PostgresURL  string
	RedisURL     string
	APIKeyPepper string
}

func Load() (Config, error) {
	listenAddr := os.Getenv("API_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}
	config := Config{
		ListenAddr:   listenAddr,
		Environment:  environment,
		PostgresURL:  os.Getenv("POSTGRES_URL"),
		RedisURL:     os.Getenv("REDIS_URL"),
		APIKeyPepper: os.Getenv("TELEMETRY_API_KEY_PEPPER"),
	}
	if environment == "production" {
		for name, value := range map[string]string{
			"POSTGRES_URL":             config.PostgresURL,
			"REDIS_URL":                config.RedisURL,
			"TELEMETRY_API_KEY_PEPPER": config.APIKeyPepper,
		} {
			if value == "" {
				return Config{}, fmt.Errorf("%s is required in production", name)
			}
		}
	}
	return config, nil
}
