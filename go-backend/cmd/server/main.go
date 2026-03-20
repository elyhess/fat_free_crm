package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/elyhess/fat-free-crm-backend/internal/config"
	"github.com/elyhess/fat-free-crm-backend/internal/database"
	"github.com/elyhess/fat-free-crm-backend/internal/handler"
)

func main() {
	cfg := config.Load()

	setupLogger(cfg.LogLevel)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	defer func() { _ = sqlDB.Close() }()

	router := handler.NewRouter(db)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}
