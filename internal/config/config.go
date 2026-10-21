package config

import (
	"os"
)

// Config holds runtime configuration settings.
type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	AdminPass string
}

// Load loads settings from the environment with robust defaults.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/confluent.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "confluent-dev-secret-key-change-in-production-12345678"
	}

	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123"
	}

	return &Config{
		Port:      port,
		DBPath:    dbPath,
		JWTSecret: jwtSecret,
		AdminPass: adminPass,
	}
}
