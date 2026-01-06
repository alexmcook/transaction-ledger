package api

import (
	"log/slog"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/twmb/franz-go/pkg/kgo"
)

/*
** Efficient JSON handler using Sonic for faster JSON parsing and VTProto for zero alloc marshalling
 */
func (s *Server) handleEfficientJSON(c fiber.Ctx) error {
	var count int // Capture number of records from the request body
	bodyPtr := trPool.Get().(*[]TransactionRequest)
	rbPtr := recordsPool.Get().(*RecordBatch)
	tempRecordBatch := false
	defer func() {
		// Clean up and return resources if not oversized
		if cap(*bodyPtr) <= 1000 {
			*bodyPtr = (*bodyPtr)[:0]
			trPool.Put(bodyPtr)
		}
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

	if err := sonic.Unmarshal(c.Body(), bodyPtr); err != nil {
		s.log.ErrorContext(c.Context(), "Invalid request body", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid request body",
		})
	}

	now := time.Now()
	body := *bodyPtr
	count = len(body)
	var records []*kgo.Record
	if count > 1000 {
		// Handle case where batch size exceeds preallocated pool size
		if count > 20000 {
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
		tx := pb.Transaction{
			Id:        body[i].ID[:],
			AccountId: body[i].AccountID[:],
			Amount:    body[i].Amount,
		}

		size := tx.SizeVT()
		payloadBuf, err := rbPtr.NextRecord(size)
		if err != nil {
			s.log.ErrorContext(c.Context(), "Failed to allocate payload buffer", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Message: "Failed to process transactions",
			})
		}

		if _, err := tx.MarshalToSizedBufferVT(payloadBuf); err != nil {
			s.log.ErrorContext(c.Context(), "Failed to marshal transaction payload", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Message: "Failed to process transactions",
			})
		}

		records[i].Topic = "transactions"
		records[i].Value = payloadBuf
		records[i].Key = body[i].AccountID[:]
		records[i].Timestamp = now
	}

	if err := s.client.ProduceSync(c.Context(), records...).FirstErr(); err != nil {
		s.log.ErrorContext(c.Context(), "Failed to sync", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to sync transactions",
		})
	}

	return c.Status(201).JSON(CreateTransactionResponse{
		CreatedCount: count,
	})
}
