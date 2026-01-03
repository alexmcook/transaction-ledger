package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (s *Server) handleGetAccount(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Invalid account ID format", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid account ID format",
		})
	}

	account, err := s.store.GetAccount(c.Context(), id)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to retrieve account", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to retrieve account",
		})
	}

	if account == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Message: "Account not found",
		})
	}

	return c.JSON(AccountResponse{
		ID:        account.ID,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
	})
}
