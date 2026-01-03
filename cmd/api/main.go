package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.NewLogger()
	log.Info("Starting Transaction Ledger API Server")

	dbUrl, ok := os.LookupEnv("DATABASE_URL_S1")
	if !ok {
		log.Error("DATABASE_URL environment variable not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("Failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	store := storage.NewPostgresStore(log, pool)
	server := api.NewServer(log, store)

	go func() {
		if err := server.Run(); err != nil {
			log.Error("Failed to start server", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("Failed to stop server", slog.String("error", err.Error()))
	}

	log.Info("Server stopped")
}
