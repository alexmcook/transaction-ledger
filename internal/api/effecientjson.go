package api

import (
	"log/slog"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/gofiber/fiber/v3"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func (s *Server) handleEfficientJSON(c fiber.Ctx) error {
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
