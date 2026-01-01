package main

import (
	"log/slog"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/logger"
)

func main() {
	log := logger.NewLogger()
	log.Info("Starting Transaction Ledger API Server")

	server := api.NewServer()
	if err := server.Run(); err != nil {
		log.Error("Failed to start server", slog.String("error", err.Error()))
	}
}
