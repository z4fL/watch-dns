package main

import (
	"log/slog"
	"os"

	"github.com/z4fL/watch-dns/internal/config"
	"github.com/z4fL/watch-dns/internal/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	_ = db

	// database.Connect(cfg.Database)
	// nextdns.NewClient(cfg.NextDNS)
	// telegram.NewBot(cfg.Telegram)
	// server.Start(cfg)

	logger.Info("server started", "port", cfg.App.Port)
}
