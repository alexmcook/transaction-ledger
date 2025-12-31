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
	"net/http"
	"os"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/alexmcook/transaction-ledger/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start metrics server: %v\n", err)
			os.Exit(1)
		}
	}()

	ctx := context.Background()

	isProd := os.Getenv("ENV") == "production"
	logger, err := logger.Init(isProd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	var maxConns int32 = 110
	pool, err := db.Connect(ctx, maxConns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		logger.ErrorContext(ctx, "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := db.NewStore(pool, logger)

	flushWorker := worker.NewFlushWorker(worker.FlushWorkerOpts{
		Logger:        logger,
		FlushInterval: 10 * 1000 * 1000 * 1000, // 10 seconds
		Flushable:     store.Transactions,
	})
	flushWorker.Start(ctx)

	txChan := make(chan *model.Transaction, 1000000)

	for i := range 10 {
		batchWorker := worker.NewBatchWorker(i, worker.BatchWorkerOpts{
			Logger:         logger,
			TxChan:         txChan,
			BatchSize:      1000,
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
		TxChan:         txChan,
	})

	server := api.NewServer(svc, logger)
	err = server.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		logger.ErrorContext(ctx, "Server failed to start", "error", err)
		os.Exit(1)
	}
}
