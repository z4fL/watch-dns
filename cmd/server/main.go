package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/z4fL/watch-dns/internal/config"
	"github.com/z4fL/watch-dns/internal/database"
	"github.com/z4fL/watch-dns/internal/handler"
	"github.com/z4fL/watch-dns/internal/logger"
	"github.com/z4fL/watch-dns/internal/nextdns"
	"github.com/z4fL/watch-dns/internal/repository"
	"github.com/z4fL/watch-dns/internal/server"
	"github.com/z4fL/watch-dns/internal/service"
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

	dnsLogRepo := repository.NewDNSLogRepository(db)
	dnsLogService := service.NewDNSLogService(
		dnsLogRepo,
	)

	nextDNSClient := nextdns.NewClient(
		cfg.NextDNS.APIKey,
	)

	ctx := context.Background()

	go func() {

		logger.Info("starting DNS log stream")

		if err := nextDNSClient.StreamLogsWithReconnect(
			ctx,
			cfg.NextDNS.ProfileID,
			func(event nextdns.LogEvent) error {
				if err := dnsLogService.Store(ctx, event.Data); err != nil {
					logger.Error(
						"failed to persist DNS log",
						"error", err,
						"event_id", event.ID,
						"domain", event.Data.Domain,
					)

					return nil
				}

				return nil
			},
		); err != nil {
			logger.Error(
				"DNS log stream stopped",
				"error", err,
			)

		}
	}()

	h := handler.New()

	srv := server.New(
		cfg.App,
		h.Router(),
	)

	logger.Info("server started", "port", cfg.App.Port)

	if err := srv.Start(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
