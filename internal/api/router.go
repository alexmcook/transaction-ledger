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
	mux.HandleFunc("/users/", handleUsers(svc))
	mux.HandleFunc("/accounts", handleAccounts(svc))
	mux.HandleFunc("/accounts/", handleAccounts(svc))
	mux.HandleFunc("/transactions", handleTransactions(svc))
	mux.HandleFunc("/transactions/", handleTransactions(svc))
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
