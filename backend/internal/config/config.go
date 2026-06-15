package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	EncryptionKey       string // exactly 32 bytes
	AppEnv              string // "development" | "production"
	Port                string
	PublicStatusEnabled bool   // enable /api/v1/* public routes + status page
	FrontendURL         string // public URL of the admin dashboard, used to link back from alerts

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
		AppEnv:              getEnvOr("APP_ENV", "development"),
		Port:                getEnvOr("PORT", "8080"),
		PublicStatusEnabled: getEnvOr("PUBLIC_STATUS_ENABLED", "true") == "true",
		FrontendURL:   strings.TrimRight(os.Getenv("FRONTEND_URL"), "/"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPass:      os.Getenv("SMTP_PASS"),
		SMTPFrom:      os.Getenv("SMTP_FROM"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.EncryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes")
	}

	smtpPortStr := getEnvOr("SMTP_PORT", "587")
	p, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return nil, fmt.Errorf("SMTP_PORT must be a number")
	}
	cfg.SMTPPort = p

	return cfg, nil
}

func (c *Config) IsProd() bool { return c.AppEnv == "production" }

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
