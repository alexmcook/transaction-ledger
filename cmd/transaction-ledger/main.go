// @title Transaction Ledger API
// @version 0.1.0
// @description Transaction Ledger API for managing users and accounts
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"context"
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"os"
)

func main() {
	ctx := context.Background()

	isProd := os.Getenv("ENV") == "PROD"
	logger, err := logger.Init(isProd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	var maxConns int32 = 200
	pool, err := db.Connect(ctx, maxConns)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := db.NewStore(pool, logger)
	svc := service.New(service.Deps{
		Logger:       logger,
		Users:        store.Users,
		Accounts:     store.Accounts,
		Transactions: store.Transactions,
	})

	server := api.NewServer(svc, logger)
	err = server.Run()
	if err != nil {
		logger.ErrorContext(ctx, "Server failed to start", "error", err)
		os.Exit(1)
	}
}
