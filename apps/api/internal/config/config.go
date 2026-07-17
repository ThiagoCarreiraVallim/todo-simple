package config

import (
	"fmt"
	"net"
	"net/url"
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

	// If DATABASE_URL isn't provided directly, build it from DB_* parts. This
	// path percent-encodes the password, so secrets containing URL-unsafe
	// characters (e.g. the "/" and "+" in base64 passwords) work correctly.
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = databaseURLFromParts()
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL (or DB_HOST/DB_USER/DB_NAME) is required")
	}

	return cfg, nil
}

// databaseURLFromParts assembles a postgres URL from discrete env vars,
// letting net/url handle the escaping. Returns "" if the required parts are
// missing.
func databaseURLFromParts() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	name := os.Getenv("DB_NAME")
	if host == "" || user == "" || name == "" {
		return ""
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, os.Getenv("DB_PASSWORD")),
		Host:   net.JoinHostPort(host, getEnv("DB_PORT", "5432")),
		Path:   "/" + name,
	}
	q := u.Query()
	q.Set("sslmode", getEnv("DB_SSLMODE", "disable"))
	u.RawQuery = q.Encode()
	return u.String()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
