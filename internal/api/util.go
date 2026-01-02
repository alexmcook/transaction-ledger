package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (s *Server) parseUUID(c fiber.Ctx, idStr string) (uuid.UUID, bool) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.log.WarnContext(c.Context(), "Invalid UUID format", slog.String("id", idStr))
		c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid UUID format",
		})
		return uuid.Nil, false
	}

	return id, true
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
