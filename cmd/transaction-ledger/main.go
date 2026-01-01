// @title			Transaction Ledger API
// @version		0.1.0
// @description	Transaction Ledger API for managing users and accounts
// @host			localhost:8080
// @BasePath		/
// @schemes		http
package main

import (
	"context"
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/alexmcook/transaction-ledger/internal/worker"
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func setupMetrics() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start metrics server: %v\n", err)
			os.Exit(1)
		}
	}()
}

func setupServer(ctx context.Context, maxConns int, logger *slog.Logger) *service.Service {
	pool, err := db.Connect(ctx, int32(maxConns))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		logger.ErrorContext(ctx, "Failed to connect to database", "error", err)
		os.Exit(1)
	}

	store := db.NewStore(pool, logger)

	flushWorker := worker.NewFlushWorker(worker.FlushWorkerOpts{
		Logger:        logger,
		FlushInterval: 10 * 1000 * 1000 * 1000, // 10 seconds
		Flushable:     store.Transactions,
	})
	flushWorker.Start(ctx)

	txChans := make([]chan model.TransactionPayload, 10)

	for i := range 10 {
		txChans[i] = make(chan model.TransactionPayload, 100000)
		batchWorker := worker.NewBatchWorker(i, worker.BatchWorkerOpts{
			Logger:         logger,
			TxChan:         txChans[i],
			BatchSize:      10000,
			BatchInterval:  250 * time.Millisecond,
			Batchable:      store.Transactions,
			BucketProvider: flushWorker,
		})
		batchWorker.Start(ctx)
	}

	svc := service.New(service.Deps{
		Logger:         logger,
		Users:          store.Users,
		Accounts:       store.Accounts,
		Transactions:   store.Transactions,
		BucketProvider: flushWorker,
		TxChans:         txChans,
		Pool:           pool,
	})

	return svc
}

func main() {
	ctx := context.Background()

	isProd := os.Getenv("ENV") == "production"
	logger, err := logger.Init(isProd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	var svc *service.Service

	// Only setup the server in child processes
	if fiber.IsChild() {
		maxConns := 110
		svc = setupServer(ctx, maxConns, logger)
	} else {
		setupMetrics()
	}

	server := api.NewServer(svc, logger)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)

		<-quit

		logger.InfoContext(ctx, "Shutting down server")

		err = server.Shutdown()
		if err != nil {
			logger.ErrorContext(ctx, "Failed to shut down server gracefully", "error", err)
		}
	}()

	err = server.Run()
	if err != nil {
		logger.ErrorContext(ctx, "Server failed to start", "error", err)
		os.Exit(1)
	}

	if fiber.IsChild() {
		logger.InfoContext(ctx, "Service: shutting down service")
		svc.Shutdown()
	}
}
