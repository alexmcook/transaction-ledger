package api

import (
	"fmt"
	"io"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/google/uuid"
)

func (s *Server) handleTransactions() http.HandlerFunc {
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
			s.handleGetTransaction()(w, r)
			return
		case http.MethodPost:
			s.handleCreateTransaction()(w, r)
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
}

func (s *Server) handleGetTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionId, err := uuid.Parse(r.PathValue("transactionId"))
		if err != nil {
			http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
			return
		}

		transaction, err := s.svc.Transactions.GetTransaction(r.Context(), transactionId)
		if err != nil {
			http.Error(w, "Transaction not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Transaction ID: %d\nAccount ID: %d\nTransaction amount: %d", transaction.Id, transaction.AccountId, transaction.Amount)
	}
}

func (s *Server) handleCreateTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Payload struct {
			AccountId uuid.UUID `json:"accountId"`
			Type      int       `json:"type"`
			Amount    int64     `json:"amount"`
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

		transaction, err := s.svc.Transactions.CreateTransaction(r.Context(), p.AccountId, p.Type, p.Amount)
		if err != nil {
			http.Error(w, "Failed to create transaction: " + err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Transaction created with ID: %d", transaction.Id)
	}
}
