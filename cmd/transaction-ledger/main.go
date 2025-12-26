package main

import (
	"context"
	"fmt"
	"os"
	"net/http"
	"github.com/joho/godotenv"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	pool, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer pool.Close()

	store := db.NewStore(pool)
	svc := service.New(service.Deps{
		Users:    store.Users,
		Accounts: store.Accounts,
	})

	httpHandler := api.NewRouter(svc)
	http.ListenAndServe(":8080", httpHandler)
}
