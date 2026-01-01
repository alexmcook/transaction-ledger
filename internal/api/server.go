package api

import (
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/gofiber/fiber/v3"
	"log/slog"
)

type Server struct {
	app    *fiber.App
	logger *slog.Logger
	svc    *service.Service
}

func NewServer(svc *service.Service, logger *slog.Logger) *Server {
	app := fiber.New()

	s := &Server{
		app:    app,
		logger: logger,
		svc:    svc,
	}

	s.registerRoutes()
	return s
}

func (s *Server) Run(addr ...string) error {
	serverAddr := ":8080"
	if len(addr) > 0 {
		serverAddr = addr[0]
	}

	if !fiber.IsChild() {
		s.logger.Info("Starting server", slog.String("addr", serverAddr))
	}

	return s.app.Listen(serverAddr, fiber.ListenConfig{
		EnablePrefork: true,
	})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) respondWithError(c fiber.Ctx, code int, message string, err error) error {
	if code >= 500 {
		s.logger.ErrorContext(c.Context(), "Server error", slog.Int("code", code), slog.String("message", message), slog.String("error", err.Error()))
	} else {
		s.logger.WarnContext(c.Context(), "Client error", slog.Int("code", code), slog.String("message", message), slog.String("error", err.Error()))
	}
	return c.Status(code).JSON(ErrorResponse{Error: message})
}

func (s *Server) json(c fiber.Ctx) error {
	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.Next()
}

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)

	api := s.app.Group("/")
	api.Use(s.json)
	api.Get("/users/:userId", s.handleGetUser)
	api.Get("/accounts/:accountId", s.handleGetAccount)
	api.Get("/transactions/:transactionId", s.handleGetTransaction)
	api.Post("/users", s.handleCreateUser)
	api.Post("/accounts", s.handleCreateAccount)
	api.Post("/transactions", s.handleCreateTransaction)
}

// @Summary		API Health check
// @Description	Returns 200 OK if the API is running
// @Produce		plain
// @Success		200	{string}	string	"OK"
// @Router			/health [get]
func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
