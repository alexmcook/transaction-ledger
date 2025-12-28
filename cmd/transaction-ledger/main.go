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
	"os"
	"github.com/joho/godotenv"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/alexmcook/transaction-ledger/internal/logger"
)

func main() {
	ctx := context.Background()

	isProd := os.Getenv("ENV") == "PROD"
	logger := logger.Init(isProd)

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	dbUrl := os.Getenv("DATABASE_URL")
	pool, err := db.Connect(ctx, dbUrl)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer pool.Close()

	store := db.NewStore(pool, logger)
	svc := service.New(service.Deps{
		Logger:  			logger,
		Users:    		store.Users,
		Accounts: 		store.Accounts,
		Transactions: store.Transactions,
	})

	server := api.NewServer(svc, logger)
	err = server.Run()
	if err != nil {
		logger.Error("Server failed to start", "error", err)
	}
}
