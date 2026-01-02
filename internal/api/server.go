package api

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

type Server struct {
	log   *slog.Logger
	app   *fiber.App
	store StoreRegistry
}

func NewServer(log *slog.Logger, store StoreRegistry) *Server {
	app := fiber.New()

	s := &Server{
		log:   log,
		app:   app,
		store: store,
	}

	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)

	s.app.Get("/users/:id", s.handleGetUser)
	s.app.Post("/users", s.handleCreateUser)

	s.app.Get("/accounts/:id", s.handleGetAccount)
	s.app.Post("/accounts", s.handleCreateAccount)

	s.app.Get("/transactions/:id", s.handleGetTransaction)
	s.app.Post("/transactions", s.handleCreateTransaction)
}

func (s *Server) Run() error {
	return s.app.Listen(":8080")
}

func (s *Server) Stop(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
