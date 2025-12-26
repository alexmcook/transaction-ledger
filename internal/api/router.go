package api

import (
	"fmt"
	"net/http"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

func NewRouter(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, svc)
	return mux
}

func registerRoutes(mux *http.ServeMux, svc *service.Service) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		user, err := svc.Users.CreateUser(r.Context())
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created with ID: %d", user.Id)
	})

	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		account, err := svc.Accounts.CreateAccount(r.Context(), 1, 1000)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Account created with ID: %d", account.Id)
	})
}
