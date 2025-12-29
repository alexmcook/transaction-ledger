package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"encoding/json"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

type Server struct {
	router 	*http.ServeMux
	logger 	*slog.Logger
	svc 		*service.Service
}

func NewServer(svc *service.Service, logger *slog.Logger) *Server {
	s := &Server{
		logger: logger,
		svc:    svc,
	}
	s.router = http.NewServeMux()
	s.registerRoutes(s.router, svc)
	return s
}

func (s *Server) Run(addr ...string) error {
	serverAddr := ":8080"
	if len(addr) > 0 {
		serverAddr = addr[0]
	}
	s.logger.Info("Starting server", slog.String("addr", serverAddr))
	return http.ListenAndServe(serverAddr, s.router)
}

func (s *Server) json(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error": "%s"}`, message)
}

func (s *Server) respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			fmt.Printf("Failed to encode JSON response: %v\n", err)
		}
	}

}

func (s *Server) registerRoutes(mux *http.ServeMux, svc *service.Service) {
	mux.Handle("/health", s.json(http.HandlerFunc(s.handleHealth)))
	mux.Handle("/users", s.json(s.handleUsers()))
	mux.Handle("/users/", s.json(s.handleUsers()))
	mux.Handle("/accounts", s.json(s.handleAccounts()))
	mux.Handle("/accounts/", s.json(s.handleAccounts()))
	mux.Handle("/transactions", s.json(s.handleTransactions()))
	mux.Handle("/transactions/", s.json(s.handleTransactions()))
}

// @Summary API Health check
// @Description Returns 200 OK if the API is running
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
