package api

import (
	"github.com/gofiber/fiber/v3"
)

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)
	s.app.Get("/accounts/:id", s.handleGetAccount)
	s.app.Get("/transactions/:id", s.handleGetTransaction)
}

func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}
