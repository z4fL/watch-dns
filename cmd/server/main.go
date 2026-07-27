package main

import (
	"log/slog"
	"os"

	"github.com/z4fL/watch-dns/internal/config"
	"github.com/z4fL/watch-dns/internal/database"
	"github.com/z4fL/watch-dns/internal/handler"
	"github.com/z4fL/watch-dns/internal/logger"
	"github.com/z4fL/watch-dns/internal/server"
	"github.com/z4fL/watch-dns/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := logger.New(cfg.Log)

	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := migrations.AutoMigrate(db); err != nil {
		logger.Error("failed to migrate models", "error", err)
		os.Exit(1)
	}

	h := handler.New()

	srv := server.New(
		cfg.App,
		h.Router(),
	)

	// nextdns.NewClient(cfg.NextDNS)
	// telegram.NewBot(cfg.Telegram)

	logger.Info("server started", "port", cfg.App.Port)

	if err := srv.Start(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
