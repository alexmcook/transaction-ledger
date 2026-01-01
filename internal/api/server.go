package api

import (
	"github.com/gofiber/fiber/v3"
	"log/slog"
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
}

func (s *Server) Run() error {
	return s.app.Listen(":8080")
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
