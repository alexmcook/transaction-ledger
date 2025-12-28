package api

import (
	"fmt"
	"io"
	"encoding/json"
	"strings"
	"net/http"
	"github.com/google/uuid"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

// @Summary Create a new account
// @Description Creates a new account for a user
// @Produce plain
// @Success 201 {string} string "Account created with ID"
// @Failure 500 {string} string "Failed to create account"
// @Router /accounts [post]
func handleAccounts(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(p, "/")
		switch r.Method {
		case http.MethodGet:
			if len(parts) != 2 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Extract account ID from URL
			r.SetPathValue("accountId", parts[1])
		  handleGetAccount(svc)(w,r)
			return
		case http.MethodPost:
			handleCreateAccount(svc)(w, r)
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
}

func handleGetAccount(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountId, err := uuid.Parse(r.PathValue("accountId"))
		if err != nil {
			http.Error(w, "Invalid account ID", http.StatusBadRequest)
			return
		}

		account, err := svc.Accounts.GetAccount(r.Context(), accountId)
		if err != nil {
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Account ID: %d\nUser ID: %d\nAccount balance: %d", account.Id, account.UserId, account.Balance)
	}
}

func handleCreateAccount(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Payload struct {
			UserId  uuid.UUID `json:"userId"`
			Balance int64     `json:"balance"`
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

		account, err := svc.Accounts.CreateAccount(r.Context(), p.UserId, p.Balance)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Account created with ID: %d", account.Id)
	}
}
