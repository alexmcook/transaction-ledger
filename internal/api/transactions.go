package api

import (
	"fmt"
	"io"
	"encoding/json"
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
		type Payload struct {
			AccountId int64 `json:"accountId"`
			Amount    int64 `json:"amount"`
		}
		var p Payload

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // limit 1MB
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(b, &p)
		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		transaction, err := svc.Transactions.CreateTransaction(r.Context(), p.AccountId, p.Amount)
		if err != nil {
			http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Transaction created with ID: %d", transaction.Id)
	}
}
