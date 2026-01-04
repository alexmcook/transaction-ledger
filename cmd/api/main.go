package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logger.NewLogger()
	log.Info("Starting Transaction Ledger API Server")

	numShards := 2
	pools := make([]*pgxpool.Pool, numShards)

	for i := range numShards {
		dbUrlEnv := fmt.Sprintf("DATABASE_URL_S%d", i+1)
		dbUrl, ok := os.LookupEnv(dbUrlEnv)
		if !ok {
			log.Error("Database URL not set", slog.String("env", dbUrlEnv))
			os.Exit(1)
		}

		pool, err := pgxpool.New(ctx, dbUrl)
		if err != nil {
			log.Error("Failed to connect to database shard", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer pool.Close()

		pools[i] = pool

		if err := pool.Ping(ctx); err != nil {
			log.Error("Failed to ping database", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_BROKERS")),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		log.Error("Failed to create broker client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx); err != nil {
		log.Error("Failed to ping broker", slog.String("error", err.Error()), slog.String("brokers", os.Getenv("KAFKA_BROKERS")))
		os.Exit(1)
	}

	shards := storage.NewShardedStore(log, pools)
	server := api.NewServer(log, shards, client)

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
