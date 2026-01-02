package api

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (s *Server) handleGetUser(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, ok := s.parseUUID(c, idStr)
	if !ok {
		return nil // parseUUID already handled the error response
	}

	user, err := s.store.Users().GetUser(c.Context(), id)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to retrieve user", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to retrieve user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Message: "User not found",
		})
	}

	return c.JSON(UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
	})
}

func (s *Server) handleCreateUser(c fiber.Ctx) error {
	id, ok := s.makeUUID(c)
	if !ok {
		return nil // makeUUID already handled the error response
	}

	now := time.Now()
	err := s.store.Users().CreateUser(c.Context(), CreateUserParams{
		ID:        id,
		CreatedAt: now,
	})
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to create user", slog.Any("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:        id,
		CreatedAt: now,
	})
}
