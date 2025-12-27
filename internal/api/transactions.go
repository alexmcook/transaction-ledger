package api

import (
	"fmt"
	"strings"
	"net/http"
	"strconv"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

func handleTransactions(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(p, "/")
		switch r.Method {
		case http.MethodGet:
			if len(parts) != 2 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Extract transaction ID from URL
			r.SetPathValue("transactionId", parts[1])
			handleGetTransaction(svc)(w, r)
			return
		case http.MethodPost:
			handleCreateTransaction(svc)(w, r)
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
}

func handleGetTransaction(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionIdStr := r.PathValue("transactionId")
		transactionId, err := strconv.ParseInt(transactionIdStr, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "Err: %s", transactionIdStr)
			http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
			return
		}

		transaction, err := svc.Transactions.GetTransaction(r.Context(), transactionId)
		if err != nil {
			http.Error(w, "Transaction not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Transaction ID: %d\nAccount ID: %d\nTransaction amount: %d", transaction.Id, transaction.AccountId, transaction.Amount)
	}
}

func handleCreateTransaction(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transaction, err := svc.Transactions.CreateTransaction(r.Context(), 1, 0)
		if err != nil {
			http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Transaction created with ID: %d", transaction.Id)
	}
}
