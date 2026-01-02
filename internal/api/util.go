package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (s *Server) parseUUID(c fiber.Ctx, param string) (uuid.UUID, bool) {
	id := c.Params(param)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.WarnContext(c.Context(), "Invalid UUID format", slog.String("id", id))
		c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid UUID format",
		})
		return uuid.Nil, false
	}

	return uid, true
}

func (s *Server) makeUUID(c fiber.Ctx) (uuid.UUID, bool) {
	id, err := uuid.NewUUID()

	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to generate UUID", slog.Any("error", err))
		c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to generate UUID",
		})
		return uuid.Nil, false
	}

	return id, true
}
