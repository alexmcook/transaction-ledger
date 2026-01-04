package api

import (
	"log/slog"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func (s *Server) handleGetTransaction(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Invalid transaction ID format", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid transaction ID format",
		})
	}

	transaction, err := s.store.GetTransaction(c.Context(), id)
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
		ID:        transaction.ID,
		AccountID: transaction.AccountID,
		Amount:    transaction.Amount,
		CreatedAt: transaction.CreatedAt,
	})
}

func (s *Server) handleCreateTransaction(c fiber.Ctx) error {
	var body []TransactionRequest
	if err := c.Bind().JSON(&body); err != nil {
		s.log.ErrorContext(c.Context(), "Invalid request body", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	s.log.InfoContext(c.Context(), "Creating transaction batch", slog.Int("count", len(body)))

	records := make([]*kgo.Record, len(body))

	for i := range body {
		payload, err := proto.Marshal(&pb.Transaction{
			Id:        body[i].ID[:],
			AccountId: body[i].AccountID[:],
			Amount:    body[i].Amount,
			CreatedAt: time.Now().UnixNano(),
		})
		if err != nil {
			s.log.ErrorContext(c.Context(), "Failed to marshal transaction payload", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Message: "Failed to process transactions",
			})
		}

		records[i] = &kgo.Record{
			Topic: "transactions",
			Value: payload,
			Key:   body[i].AccountID[:],
		}
	}

	if err := s.client.ProduceSync(c.Context(), records...).FirstErr(); err != nil {
		s.log.ErrorContext(c.Context(), "Failed to sync", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to sync transactions",
		})
	}

	return c.JSON(CreateTransactionResponse{
		CreatedCount: len(body),
	})
}
