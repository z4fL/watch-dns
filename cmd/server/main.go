package main

import (
	"log/slog"
	"os"

	"github.com/z4fL/watch-dns/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info(
		"configuration loaded",
		"app", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// database.Connect(cfg.Database)
	// nextdns.NewClient(cfg.NextDNS)
	// telegram.NewBot(cfg.Telegram)
	// server.Start(cfg)

	logger.Info("server started", "port", cfg.App.Port)
}
