package api

import (
	"github.com/gofiber/fiber/v3"
)

type Server struct {
	app *fiber.App
}

func NewServer() *Server {
	app := fiber.New()
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})
	return &Server{
		app: app,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(":8080")
}
