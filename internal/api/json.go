package api

import (
	"log/slog"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/gofiber/fiber/v3"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func (s *Server) handleJSON(c fiber.Ctx) error {
	var body []TransactionRequest

	unmarshalStart := time.Now()
	if err := c.Bind().JSON(&body); err != nil {
		s.log.ErrorContext(c.Context(), "Invalid request body", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}
	unmarshalLatency.WithLabelValues("json").Observe(time.Since(unmarshalStart).Seconds())

	s.log.DebugContext(c.Context(), "Creating transaction batch", slog.Int("count", len(body)))

	records := make([]*kgo.Record, len(body))

	for i := range body {
		payload, err := proto.Marshal(&pb.Transaction{
			Id:        body[i].ID[:],
			AccountId: body[i].AccountID[:],
			Amount:    body[i].Amount,
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

	kafkaStart := time.Now()
	if err := s.client.ProduceSync(c.Context(), records...).FirstErr(); err != nil {
		s.log.ErrorContext(c.Context(), "Failed to sync", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to sync transactions",
		})
	}
	kafkaProducerLatency.Observe(time.Since(kafkaStart).Seconds())
	kafkaTransactionsProduced.Add(float64(len(body)))

	return c.Status(201).JSON(CreateTransactionResponse{
		CreatedCount: len(body),
	})
}
