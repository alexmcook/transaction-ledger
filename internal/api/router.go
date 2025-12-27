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
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/users", handleUsers(svc))
	mux.HandleFunc("/accounts", handleAccounts(svc))
}

// @Summary API Health check
// @Description Returns 200 OK if the API is running
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// @Summary Create a new user
// @Description Creates a new user in the system
// @Produce plain
// @Success 201 {string} string "User created with ID"
// @Failure 500 {string} string "Failed to create user"
// @Router /users [post]
func handleUsers(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := svc.Users.CreateUser(r.Context())
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created with ID: %d", user.Id)
	}
}

// @Summary Create a new account
// @Description Creates a new account for a user
// @Produce plain
// @Success 201 {string} string "Account created with ID"
// @Failure 500 {string} string "Failed to create account"
// @Router /accounts [post]
func handleAccounts(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := svc.Accounts.CreateAccount(r.Context(), 1, 1000)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Account created with ID: %d", account.Id)
	}
}
