package api

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (s *Server) handleGetAccount(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, ok := s.parseUUID(c, idStr)
	if !ok {
		return nil // parseUUID already handled the error response
	}

	account, err := s.store.Accounts().GetAccount(c.Context(), id)
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
		UserID:    account.UserID,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
	})
}

func (s *Server) handleCreateAccount(c fiber.Ctx) error {
	id, ok := s.makeUUID(c)
	if !ok {
		return nil // makeUUID already handled the error response
	}
	now := time.Now()

	var req CreateAccountRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	err := s.store.Accounts().CreateAccount(c.Context(), CreateAccountParams{
		ID:        id,
		UserID:    req.UserID,
		Balance:   req.Balance,
		CreatedAt: now,
	})
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to create account", slog.Any("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create account",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(AccountResponse{
		ID:        id,
		UserID:    req.UserID,
		Balance:   req.Balance,
		CreatedAt: now,
	})
}
