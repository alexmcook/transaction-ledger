package api

import (
	"context"
	"encoding/json"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"log/slog"
	"net/http"
)

type Server struct {
	router *http.ServeMux
	logger *slog.Logger
	svc    *service.Service
}

func NewServer(svc *service.Service, logger *slog.Logger) *Server {
	s := &Server{
		logger: logger,
		svc:    svc,
	}
	s.router = http.NewServeMux()
	s.registerRoutes(s.router)
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

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) respondWithError(ctx context.Context, w http.ResponseWriter, code int, message string, err error) {
	if code >= 500 {
		s.logger.ErrorContext(ctx, "Server error", slog.Int("code", code), slog.String("message", message), slog.String("error", err.Error()))
	} else if code >= 400 {
		s.logger.WarnContext(ctx, "Client error", slog.Int("code", code), slog.String("message", message), slog.String("error", err.Error()))
	}
	w.WriteHeader(code)
	errResp := ErrorResponse{Error: message}
	err = json.NewEncoder(w).Encode(errResp)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to encode error response", slog.String("error", err.Error()))
	}
}

func (s *Server) respondWithJSON(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	w.WriteHeader(code)
	if payload == nil {
		s.logger.DebugContext(ctx, "Response", slog.Int("code", code))
		return
	}
	s.logger.DebugContext(ctx, "Response", slog.Int("code", code), slog.Any("payload", payload))
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to encode JSON response", slog.String("error", err.Error()))
	}
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.Handle("/health", http.HandlerFunc(s.handleHealth))
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}
