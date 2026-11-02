package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

const (
	defaultPort        = "8080"
	defaultDatabaseURL = "postgres://chatuser:chatpass@localhost:5432/chatapp?sslmode=disable"
	defaultJWTSecret   = "dev-secret-change-me"
	defaultOrigin      = "http://localhost:3000"

	minJWTSecretLength = 32
)

type Config struct {
	// Env is "development" (default) or "production"
	Env            string
	Port           string
	DatabaseURL    string
	JWTSecret      string
	AllowedOrigins []string
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// Load reads settings from the environment. In development, missing values
// fall back to local defaults; in production, they must be set explicitly.
func Load() (*Config, error) {
	cfg := &Config{
		Env:         strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))),
		Port:        strings.TrimSpace(os.Getenv("PORT")),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.Env != "development" && cfg.Env != "production" {
		return nil, fmt.Errorf("APP_ENV must be \"development\" or \"production\", got %q", cfg.Env)
	}

	if cfg.Port == "" {
		cfg.Port = defaultPort
	}

	for _, origin := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin != "" {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
		}
	}

	if err := cfg.applyDefaults(); err != nil {
		return nil, err
	}
	if err := validateDatabaseURL(cfg.DatabaseURL); err != nil {
		return nil, err
	}
	for _, origin := range cfg.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (c *Config) applyDefaults() error {
	if c.IsProduction() {
		var missing []string
		if c.DatabaseURL == "" {
			missing = append(missing, "DATABASE_URL")
		}
		if c.JWTSecret == "" {
			missing = append(missing, "JWT_SECRET")
		}
		if len(c.AllowedOrigins) == 0 {
			missing = append(missing, "ALLOWED_ORIGINS")
		}
		if len(missing) > 0 {
			return fmt.Errorf("production requires %s to be set", strings.Join(missing, ", "))
		}
		if c.JWTSecret == defaultJWTSecret || len(c.JWTSecret) < minJWTSecretLength {
			return fmt.Errorf("JWT_SECRET must be a random value of at least %d characters in production", minJWTSecretLength)
		}
		return nil
	}

	if c.DatabaseURL == "" {
		c.DatabaseURL = defaultDatabaseURL
	}
	if c.JWTSecret == "" {
		log.Printf("JWT_SECRET not set; using an insecure development secret")
		c.JWTSecret = defaultJWTSecret
	}
	if len(c.AllowedOrigins) == 0 {
		c.AllowedOrigins = []string{defaultOrigin}
	}
	return nil
}

func validateDatabaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DATABASE_URL is not a valid URL: %w", err)
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must start with postgres:// or postgresql://")
	}
	if u.Host == "" {
		return fmt.Errorf("DATABASE_URL is missing a host")
	}
	if strings.TrimPrefix(u.Path, "/") == "" {
		return fmt.Errorf("DATABASE_URL is missing a database name")
	}
	return nil
}

func validateOrigin(origin string) error {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("ALLOWED_ORIGINS entry %q must look like https://example.com", origin)
	}
	return nil
}
