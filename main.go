package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-transaction-service/api"
	"inventory-transaction-service/config"
	"inventory-transaction-service/repository"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.LoadConfig(".")
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dynamoClient, err := repository.NewDynamoDBClient(ctx, cfg)
	if err != nil {
		slog.Error("Failed to build DynamoDB client", slog.Any("error", err))
		os.Exit(1)
	}

	repo := repository.NewTransactionRepository(dynamoClient, cfg.DynamoDBTableName)
	if err := repo.EnsureTable(ctx); err != nil {
		// Don't crash on table provisioning errors against real AWS — log and
		// continue so the service still serves traffic to a pre-existing table.
		slog.Error("Could not ensure DynamoDB table", slog.Any("error", err))
	}

	srv, err := api.NewServer(ctx, cfg, repo)
	if err != nil {
		slog.Error("Failed to build server", slog.Any("error", err))
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := net.JoinHostPort("", cfg.ServerPort)
		if err := srv.Run(addr); err != nil {
			slog.Error("HTTP server failed", slog.Any("error", err))
			cancel()
		}
	}()

	<-quit
	slog.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", slog.Any("error", err))
	}
	slog.Info("Server exited")
}
