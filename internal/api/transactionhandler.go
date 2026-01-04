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

	batch := &pb.TransactionBatch{
		Transactions: make([]*pb.Transaction, 0, len(body)),
	}

	for _, tr := range body {
		batch.Transactions = append(batch.Transactions, &pb.Transaction{
			Id:        tr.ID[:],
			AccountId: tr.AccountID[:],
			Amount:    tr.Amount,
			CreatedAt: time.Now().UnixNano(),
		})
	}

	payload, err := proto.Marshal(batch)
	if err != nil {
		s.log.ErrorContext(c.Context(), "Failed to marshal transaction batch", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to process transactions",
		})
	}

	results := s.client.ProduceSync(c.Context(), &kgo.Record{
		Topic: "transactions",
		Value: payload,
	})

	if err := results.FirstErr(); err != nil {
		s.log.ErrorContext(c.Context(), "Failed to produce transaction batch to broker", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to process transactions",
		})
	}

	return c.JSON(CreateTransactionResponse{
		CreatedCount: len(batch.Transactions),
	})
}
