package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/alexmcook/transaction-ledger/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

func setup() (*worker.Writer, func(), error) {
	var closures []func()
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			for i := len(closures) - 1; i >= 0; i-- {
				closures[i]()
			}
		})
	}

	log := logger.NewLogger(slog.LevelDebug)
	log.Info("Starting transaction ledger worker")

	numShards, err := strconv.Atoi(os.Getenv("NUM_SHARDS"))
	if err != nil || numShards <= 0 {
		return nil, cleanup, fmt.Errorf("invalid NUM_SHARDS value: %v", os.Getenv("NUM_SHARDS"))
	}

	pools := make([]*pgxpool.Pool, numShards)

	for i := range numShards {
		dbUrlEnv, ok := os.LookupEnv("DATABASE_URL")
		if !ok {
			return nil, cleanup, fmt.Errorf("DATABASE_URL environment variable not set")
		}

		dbUrl := fmt.Sprintf(dbUrlEnv, i+1) // postgres-s%d

		pool, err := pgxpool.New(context.Background(), dbUrl)
		if err != nil {
			return nil, cleanup, fmt.Errorf("failed to connect to database shard: %v", err)
		}
		closures = append(closures, pool.Close)

		pools[i] = pool

		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err != nil {
			return nil, cleanup, fmt.Errorf("failed to ping database: %v", err)
		}
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_BROKERS")),
		kgo.ConsumerGroup("transaction-writer"),
		kgo.ConsumeTopics("transactions"),
	)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create broker client: %v", err)
	}
	closures = append(closures, client.Close)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Ping(pingCtx)
	cancel()
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to ping broker client: %v", err)
	}

	shards := storage.NewShardedStore(log, pools)
	writer := worker.NewWriter(log, shards, client, numShards)

	return writer, cleanup, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start pprof server: %v\n", err)
			os.Exit(1)
		}
	}()

	writer, cleanup, err := setup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set up server: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start worker: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := writer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop worker: %v\n", err)
	}

	cleanup()
	fmt.Println("Server stopped")
}
