package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that would override defaults
	for _, key := range []string{"PORT", "DATABASE_URL", "LOG_LEVEL", "DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE"} {
		t.Setenv(key, "")
	}

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log level info, got %s", cfg.LogLevel)
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected non-empty database URL")
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://custom:5432/testdb")
	t.Setenv("LOG_LEVEL", "debug")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://custom:5432/testdb" {
		t.Errorf("expected custom DATABASE_URL, got %s", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
}

func TestBuildDefaultDSN(t *testing.T) {
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE", "DB_SSLMODE"} {
		t.Setenv(key, "")
	}

	dsn := buildDefaultDSN()

	expected := "host=localhost port=5432 dbname=fat_free_crm_development sslmode=disable"
	if dsn != expected {
		t.Errorf("expected %q, got %q", expected, dsn)
	}
}

func TestBuildDefaultDSN_WithCredentials(t *testing.T) {
	t.Setenv("DB_USERNAME", "myuser")
	t.Setenv("DB_PASSWORD", "mypass")
	t.Setenv("DB_DATABASE", "mydb")

	dsn := buildDefaultDSN()

	if dsn != "host=localhost port=5432 dbname=mydb sslmode=disable user=myuser password=mypass" {
		t.Errorf("unexpected DSN: %s", dsn)
	}
}
