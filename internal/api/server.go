package api

import (
	"github.com/gofiber/fiber/v3"
)

type Server struct {
	app   *fiber.App
	store StoreRegistry
}

func NewServer(store StoreRegistry) *Server {
	app := fiber.New()

	s := &Server{
		app:   app,
		store: store,
	}

	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	s.app.Get("/health", handleHealth)
	s.app.Get("/users/:id", handleGetUser)
	s.app.Post("/users", handleCreateUser)
}

func (s *Server) Run() error {
	return s.app.Listen(":8080")
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
