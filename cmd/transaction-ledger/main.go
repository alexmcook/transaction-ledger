package main

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()

	log := logger.NewLogger()
	log.Info("Starting Transaction Ledger API Server")

	if err := godotenv.Load(); err != nil {
		log.Error("No .env file found")
		os.Exit(1)
	}

	dbUrl, ok := os.LookupEnv("DATABASE_URL")
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

	store := storage.NewPostgresStore(pool)

	server := api.NewServer(log, store)
	if err := server.Run(); err != nil {
		log.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
