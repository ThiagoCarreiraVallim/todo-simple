package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application's runtime configuration, read from the
// environment. In development, values fall back to the monorepo root .env
// (two directories up from apps/api) so a single .env file works for every
// app in the workspace.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	WebOrigin   string
}

func Load() (Config, error) {
	// Best-effort: ignore a missing file (e.g. in prod, where env vars are
	// injected directly by the platform).
	_ = godotenv.Load("../../.env")

	cfg := Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("API_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		WebOrigin:   getEnv("WEB_ORIGIN", "http://localhost:3000"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
