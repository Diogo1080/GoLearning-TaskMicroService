package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Port                string
	IdentityServiceAddr string
	Database            DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load(envFile string) (Config, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil && !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("load environment file: %w", err)
		}
	}

	cfg := Config{
		AppEnv:              valueOrDefault("APP_ENV", "development"),
		Port:                valueOrDefault("PORT", "9002"),
		IdentityServiceAddr: firstValue("IDENTITY_SERVICE_ADDR", "IDENTITY-SERVICE-ADDR"),
		Database: DatabaseConfig{
			Host:     valueOrDefault("DB_HOST", "localhost"),
			Port:     valueOrDefault("DB_PORT", "5432"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_SECRET"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  valueOrDefault("DB_SSLMODE", "disable"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a valid TCP port")
	}
	if strings.TrimSpace(c.IdentityServiceAddr) == "" {
		return fmt.Errorf("IDENTITY_SERVICE_ADDR is required")
	}
	for name, value := range map[string]string{
		"DB_USER":   c.Database.User,
		"DB_SECRET": c.Database.Password,
		"DB_NAME":   c.Database.Name,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if len(c.Database.Password) < 4 {
		return fmt.Errorf("DB_SECRET must be at least 4 characters")
	}
	return nil
}

func (c Config) DatabaseURL() string {
	databaseURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Database.User, c.Database.Password),
		Host:   c.Database.Host + ":" + c.Database.Port,
		Path:   "/" + c.Database.Name,
	}
	query := databaseURL.Query()
	query.Set("sslmode", c.Database.SSLMode)
	databaseURL.RawQuery = query.Encode()
	return databaseURL.String()
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func firstValue(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
