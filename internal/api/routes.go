package api

import (
	"github.com/gofiber/fiber/v3"
)

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.handleHealth)
	s.app.Get("/accounts/:id", s.handleGetAccount)
	s.app.Get("/transactions/:id", s.handleGetTransaction)

	s.app.Post("/transactions/json", s.handleJSON)
	s.app.Post("/transactions/effjson", s.handleEfficientJSON)
	s.app.Post("/transactions/proto", s.handleProto)
}

func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}
