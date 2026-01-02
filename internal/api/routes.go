package api

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (s *Server) handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}

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

func (s *Server) handleGetTransaction(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, ok := s.parseUUID(c, idStr)
	if !ok {
		return nil // parseUUID already handled the error response
	}

	transaction, err := s.store.Transactions().GetTransaction(c.Context(), id)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to retrieve transaction", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to retrieve transaction",
		})
	}

	if transaction == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Message: "Transaction not found",
		})
	}

	return c.JSON(TransactionResponse{
		ID:              transaction.ID,
		AccountID:       transaction.AccountID,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		CreatedAt:       transaction.CreatedAt,
	})
}

func (s *Server) handleCreateTransaction(c fiber.Ctx) error {
	id, ok := s.makeUUID(c)
	if !ok {
		return nil // makeUUID already handled the error response
	}
	now := time.Now()

	var req CreateTransactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	err := s.store.Transactions().CreateTransaction(c.Context(), CreateTransactionParams{
		ID:              id,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		CreatedAt:       now,
	})
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to create transaction", slog.Any("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create transaction",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(TransactionResponse{
		ID:              id,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		CreatedAt:       now,
	})
}
