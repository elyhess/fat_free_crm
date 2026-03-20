package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", buildDefaultDSN()),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func buildDefaultDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USERNAME", "")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_DATABASE", "fat_free_crm_development")
	sslMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s", host, port, dbName, sslMode)
	if user != "" {
		dsn += fmt.Sprintf(" user=%s", user)
	}
	if password != "" {
		dsn += fmt.Sprintf(" password=%s", password)
	}
	return dsn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
