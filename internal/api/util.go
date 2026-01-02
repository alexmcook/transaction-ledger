package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"log/slog"
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
