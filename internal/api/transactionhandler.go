package api

import (
	"log/slog"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

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
	var req CreateTransactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	err := s.store.Transactions().CreateTransaction(c.Context(), req)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to create transaction", slog.Any("account_id", req.AccountID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create transaction",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(SingleTransactionResponse{
		CreatedCount: 1,
	})
}

var txSlicePool = sync.Pool{
	New: func() any {
		s := make([]CreateTransactionRequest, 0, 100)
		return &s
	},
}

func getTxSlice() *[]CreateTransactionRequest {
	return txSlicePool.Get().(*[]CreateTransactionRequest)
}

func putTxSlice(s *[]CreateTransactionRequest) {
	*s = (*s)[:0]
	txSlicePool.Put(s)
}

func (s *Server) handleCreateBatchTransaction(c fiber.Ctx) error {
	req := getTxSlice()
	defer putTxSlice(req)

	if err := sonic.Unmarshal(c.Body(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	count, err := s.store.Transactions().CreateBinaryBatchTransaction(c.Context(), *req)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to create batch transactions", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to create batch transactions",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(BatchTransactionResponse{
		CreatedCount: count,
	})
}
