package config

import (
	"fmt"
	"os"
	"time"
)

type ServerTimeouts struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DependencyTimeouts struct {
	PostgresTimeout time.Duration
	RedisTimeout    time.Duration
}

type Config struct {
	ListenAddr      string
	Environment     string
	PostgresURL     string
	RedisURL        string
	APIKeyPepper    string
	Server          ServerTimeouts
	Dependencies    DependencyTimeouts
	ShutdownTimeout time.Duration
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
	serverTimeouts, err := loadServerTimeouts()
	if err != nil {
		return Config{}, err
	}
	dependencyTimeouts, err := loadDependencyTimeouts()
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := loadDuration("API_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	config := Config{
		ListenAddr:      listenAddr,
		Environment:     environment,
		PostgresURL:     os.Getenv("POSTGRES_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		APIKeyPepper:    os.Getenv("TELEMETRY_API_KEY_PEPPER"),
		Server:          serverTimeouts,
		Dependencies:    dependencyTimeouts,
		ShutdownTimeout: shutdownTimeout,
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

func loadServerTimeouts() (ServerTimeouts, error) {
	readHeaderTimeout, err := loadDuration("API_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return ServerTimeouts{}, err
	}
	readTimeout, err := loadDuration("API_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return ServerTimeouts{}, err
	}
	writeTimeout, err := loadDuration("API_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return ServerTimeouts{}, err
	}
	idleTimeout, err := loadDuration("API_IDLE_TIMEOUT", time.Minute)
	if err != nil {
		return ServerTimeouts{}, err
	}
	return ServerTimeouts{
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}, nil
}

func loadDependencyTimeouts() (DependencyTimeouts, error) {
	postgresTimeout, err := loadDuration("API_POSTGRES_TIMEOUT", 2*time.Second)
	if err != nil {
		return DependencyTimeouts{}, err
	}
	redisTimeout, err := loadDuration("API_REDIS_TIMEOUT", 2*time.Second)
	if err != nil {
		return DependencyTimeouts{}, err
	}
	return DependencyTimeouts{PostgresTimeout: postgresTimeout, RedisTimeout: redisTimeout}, nil
}

func loadDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return duration, nil
}
