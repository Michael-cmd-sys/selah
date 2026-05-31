package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type GoogleOAuth struct {
	ClientID     string
	ClientSecret string
}

type Config struct {
	DatabaseURL string
	ListenAddr  string
	JWTSecret   string
	JWTExpiry   time.Duration
	Google      GoogleOAuth
	AppEnv      string
}

func Load() (*Config, error) {
	dbURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	expiryHours := getEnv("JWT_EXPIRY_HOURS", "720")
	hours, err := strconv.Atoi(expiryHours)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}

	return &Config{
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		JWTExpiry:   time.Duration(hours) * time.Hour,
		ListenAddr:  getEnv("LISTEN_ADDR", ":8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		Google: GoogleOAuth{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		},
	}, nil
}

func (c *Config) IsDev() bool { return c.AppEnv == "development" }

func requireEnv(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("%s is required", key)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
