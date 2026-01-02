package api

import (
	"github.com/gofiber/fiber/v3"
)

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)

	s.app.Get("/users/:id", s.handleGetUser)
	s.app.Post("/users", s.handleCreateUser)

	s.app.Get("/accounts/:id", s.handleGetAccount)
	s.app.Post("/accounts", s.handleCreateAccount)

	s.app.Get("/transactions/:id", s.handleGetTransaction)
	s.app.Post("/transactions", s.handleCreateTransaction)
}

func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}
