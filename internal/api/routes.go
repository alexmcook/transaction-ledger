package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}

func (s *Server) handleGetUser(c fiber.Ctx) error {
	id := c.Params("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid UUID format",
		})
	}

	user, err := s.store.Users().GetUser(c.Context(), uid)
	if err != nil {
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
	id, err := uuid.NewV7()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to generate UUID",
		})
	}

	now := time.Now()
	err = s.store.Users().CreateUser(c.Context(), CreateUserParams{
		ID:        id,
		CreatedAt: now,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:        id,
		CreatedAt: now,
	})
}

func (s *Server) handleGetAccount(c fiber.Ctx) error {
	id := c.Params("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid UUID format",
		})
	}

	account, err := s.store.Accounts().GetAccount(c.Context(), uid)
	if err != nil {
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
	id, err := uuid.NewV7()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to generate UUID",
		})
	}
	now := time.Now()

	var req CreateAccountRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	err = s.store.Accounts().CreateAccount(c.Context(), CreateAccountParams{
		ID:        id,
		UserID:    req.UserID,
		Balance:   req.Balance,
		CreatedAt: now,
	})
	if err != nil {
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
