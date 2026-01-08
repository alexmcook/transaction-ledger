package api

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Server struct {
	log    *slog.Logger
	app    *fiber.App
	store  StoreRegistry
	client *kgo.Client
}

func NewServer(log *slog.Logger, store StoreRegistry, client *kgo.Client) *Server {
	app := fiber.New()
	app.Use(pprof.New())

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Use(prometheusMiddleware)

	s := &Server{
		log:    log,
		app:    app,
		store:  store,
		client: client,
	}

	s.registerRoutes()
	return s
}

func (s *Server) Run() error {
	return s.app.Listen(":8080")
}

func (s *Server) Stop(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
