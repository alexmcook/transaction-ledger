package api

import (
	"log/slog"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/gofiber/fiber/v3"
	"github.com/twmb/franz-go/pkg/kgo"
)

/*
** Efficient Protobuf handler
 */
func (s *Server) handleProto(c fiber.Ctx) error {
	batch := txBatchPool.Get().(*pb.TransactionBatch)
	rbPtr := recordsPool.Get().(*RecordBatch)

	var count int
	tempRecordBatch := false

	defer func() {
		// Clean up and return resources if not oversized
		batch.Transactions = batch.Transactions[:0]
		txBatchPool.Put(batch)

		if !tempRecordBatch {
			rbPtr.Reset(count)
			recordsPool.Put(rbPtr)
		}
	}()

	if len(c.Body()) > 100000*128 {
		s.log.ErrorContext(c.Context(), "Request body too large", slog.Int("size", len(c.Body())))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Request body too large",
		})
	}

	unmarhshalStart := time.Now()
	if err := batch.UnmarshalVT(c.Body()); err != nil {
		s.log.ErrorContext(c.Context(), "Invalid request body", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}
	unmarshalLatency.WithLabelValues("protobuf").Observe(time.Since(unmarhshalStart).Seconds())

	now := time.Now()
	body := batch.Transactions
	count = len(body)
	var records []*kgo.Record
	if count > 1000 {
		// Handle case where batch size exceeds preallocated pool size
		if count > 10000 {
			s.log.ErrorContext(c.Context(), "Request batch size too large", slog.Int("count", count))
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Message: "Request batch size too large",
			})
		}
		s.log.WarnContext(c.Context(), "Request batch size exceeds pool size, allocating temporary buffers", slog.Int("count", count))
		tempRecordBatch = true
		rbPtr = &RecordBatch{
			Slab:     make([]kgo.Record, count),
			Pointers: make([]*kgo.Record, count),
			ByteSlab: make([]byte, count*128), // Allocate byte slab with a safety margin
			offset:   0,
		}
		for i := range rbPtr.Slab {
			rbPtr.Pointers[i] = &rbPtr.Slab[i]
		}
	}
	records = rbPtr.Pointers[:count] // Slice pointers to the number of records

	s.log.DebugContext(c.Context(), "Creating transaction batch", slog.Int("count", count))

	for i := range count {
		size := body[i].SizeVT()
		payloadBuf, err := rbPtr.NextRecord(size)
		if err != nil {
			s.log.ErrorContext(c.Context(), "Failed to allocate payload buffer", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Message: "Failed to process transactions",
			})
		}

		if _, err := body[i].MarshalToSizedBufferVT(payloadBuf); err != nil {
			s.log.ErrorContext(c.Context(), "Failed to marshal transaction payload", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Message: "Failed to process transactions",
			})
		}

		records[i].Topic = "transactions"
		records[i].Value = payloadBuf
		records[i].Key = body[i].AccountId[:]
		records[i].Timestamp = now
	}

	kafkaStart := time.Now()
	if err := s.client.ProduceSync(c.Context(), records...).FirstErr(); err != nil {
		s.log.ErrorContext(c.Context(), "Failed to sync", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to sync transactions",
		})
	}
	kafkaProducerLatency.Observe(time.Since(kafkaStart).Seconds())
	kafkaTransactionsProduced.Add(float64(count))

	return c.Status(201).JSON(CreateTransactionResponse{
		CreatedCount: count,
	})
}
