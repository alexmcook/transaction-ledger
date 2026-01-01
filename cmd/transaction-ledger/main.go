package main

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()

	log := logger.NewLogger()
	log.Info("Starting Transaction Ledger API Server")

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Error("Failed to connect to database", slog.String("error", err.Error()))
		return
	}
	defer pool.Close()

	store := storage.NewPostgresStore(pool)

	server := api.NewServer(store)
	if err := server.Run(); err != nil {
		log.Error("Failed to start server", slog.String("error", err.Error()))
	}
}
