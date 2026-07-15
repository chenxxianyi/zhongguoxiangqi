package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	EngineMoveTime  time.Duration
	MaxUploadBytes  int64
	AllowedOrigin   string
	DataMode        string
}

func Load() (Config, error) {
	shutdownTimeout, err := durationEnv("XIANGQI_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	engineMoveTime, err := durationEnv("XIANGQI_ENGINE_MOVE_TIME", 600*time.Millisecond)
	if err != nil {
		return Config{}, err
	}
	maxUploadBytes, err := int64Env("XIANGQI_MAX_UPLOAD_BYTES", 2<<20)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		HTTPAddr:        env("XIANGQI_HTTP_ADDR", ":8080"),
		ShutdownTimeout: shutdownTimeout,
		EngineMoveTime:  engineMoveTime,
		MaxUploadBytes:  maxUploadBytes,
		AllowedOrigin:   env("XIANGQI_ALLOWED_ORIGIN", "http://localhost:5666"),
		DataMode:        env("XIANGQI_DATA_MODE", "memory"),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("XIANGQI_HTTP_ADDR must not be empty")
	}
	if cfg.ShutdownTimeout <= 0 || cfg.EngineMoveTime <= 0 {
		return Config{}, fmt.Errorf("timeouts must be positive")
	}
	if cfg.MaxUploadBytes < 1024 || cfg.MaxUploadBytes > 64<<20 {
		return Config{}, fmt.Errorf("XIANGQI_MAX_UPLOAD_BYTES must be between 1 KiB and 64 MiB")
	}
	if cfg.DataMode != "memory" {
		return Config{}, fmt.Errorf("unsupported XIANGQI_DATA_MODE %q; this build supports memory", cfg.DataMode)
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}

func int64Env(key string, fallback int64) (int64, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}
