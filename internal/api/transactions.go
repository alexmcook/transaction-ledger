package api

import (
	"github.com/alexmcook/transaction-ledger/internal/metrics"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

// TransactionResponse represents the transaction data returned in API responses
type TransactionResponse struct {
	// Id is the unique identifier of the transaction
	//	@example	770e8400-e29b-41d4-a716-446655440000
	Id uuid.UUID `json:"id"`
	// AccountId is the unique identifier of the account associated with the transaction
	//	@example	880e8400-e29b-41d4-a716-446655440000
	AccountId uuid.UUID `json:"accountId"`
	// Amount is the amount of the transaction
	//	@example	500
	Amount int64 `json:"amount"`
	// Type is the type of the transaction (e.g., credit or debit)
	//	@example	0
	Type int `json:"type"`
	// CreatedAt is the timestamp when the transaction was created
	//	@example	2025-12-25T11:11:00Z
	CreatedAt time.Time `json:"createdAt"`
}

func toTransactionResponse(t *model.Transaction) *TransactionResponse {
	return &TransactionResponse{
		Id:        t.Id,
		AccountId: t.AccountId,
		Amount:    t.Amount,
		Type:      t.Type,
		CreatedAt: time.UnixMilli(t.CreatedAt),
	}
}

// @Summary		Get transaction
// @Description	Retrieves a transaction by its ID
// @Produce		json
// @Param			transactionId	path		string				true	"Transaction ID"	format(uuid)
// @Success		200				{object}	TransactionResponse	"Transaction object"
// @Failure		400				{object}	ErrorResponse		"Invalid transaction ID"
// @Failure		404				{object}	ErrorResponse		"Transaction not found"
// @Router			/transactions/{transactionId} [get]
func (s *Server) handleGetTransaction(c fiber.Ctx) error {
	var params struct {
		TransactionId string `params:"transactionId"`
	}

	err := c.Bind().URI(&params)
	if err != nil {
		return s.respondWithError(c, fiber.StatusBadRequest, "Invalid request parameters", err)
	}

	transactionId, err := uuid.Parse(params.TransactionId)
	if err != nil {
		return s.respondWithError(c, fiber.StatusBadRequest, "Invalid transaction ID format", err)
	}

	transaction, err := s.svc.Transactions.GetTransaction(c.Context(), transactionId)
	if err != nil {
		return s.respondWithError(c, fiber.StatusNotFound, "Transaction not found", err)
	}

	return c.JSON(toTransactionResponse(transaction))
}

// @Summary		Create a new transaction
// @Description	Creates a new transaction for an account
// @Produce		json
// @Param			transaction	body		model.TransactionPayload	true	"Transaction payload"
// @Success		201			{object}	TransactionResponse	"Transaction object"
// @Failure		400			{object}	ErrorResponse		"Invalid request payload"
// @Failure		500			{object}	ErrorResponse		"Failed to create transaction"
// @Router			/transactions [post]
func (s *Server) handleCreateTransaction(c fiber.Ctx) error {
	var p model.TransactionPayload

	err := c.Bind().Body(&p)
	if err != nil {
		return s.respondWithError(c, fiber.StatusBadRequest, "Invalid JSON payload", err)
	}

	workerIdx := getWorkerID(p.AccountId, len(s.svc.TxChans))

	select {
	case s.svc.TxChans[workerIdx] <- p:
		metrics.TransactionsSuccess.Inc()
	default:
		// Channel is full, respond with service unavailable
		return s.respondWithError(c, fiber.StatusServiceUnavailable, "Transaction service is busy", nil)
	}

	return c.SendStatus(fiber.StatusCreated)
}

func getWorkerID(accountId uuid.UUID, numWorkers int) int {
	const (
		offset32 = 2166136261
		prime32 = 16777619
	)
	hash := uint32(offset32)
	for _, b := range accountId {
		hash ^= uint32(b)
		hash *= prime32
	}
	return int(hash % uint32(numWorkers))
}
