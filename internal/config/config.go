package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/* config.go -> Loads Application's configurations from environment variables
and throws an error when those values are not provided*/

// Config struct holds all configuration values an application needs
type Config struct {
	DatabaseURL, ListenAddr, OTLPEndpoint string
	ShutdownTimeout                       time.Duration //maximum time allowed for graceful shutdown
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ListenAddr:      os.Getenv("LISTEN_ADDR"),
		OTLPEndpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		ShutdownTimeout: 15 * time.Second,
	}

	// the application must recieve one through an environment variable(fail first config)
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}

	// checks if SHUTDOWN_TIMEOUT_SECONDS environment variable exist
	if shutdownVar := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); shutdownVar != "" {
		shutdownVarInt, err := strconv.Atoi(shutdownVar)

		if err != nil || shutdownVarInt < 1 || shutdownVarInt > 120 {
			return cfg, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS")
		}
	}

	return cfg, nil
}
