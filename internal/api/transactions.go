package api

import (
	"encoding/json"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"net/http"
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

// TransactionPayload represents the transaction data received in API requests
type TransactionPayload struct {
	// AccountId is the unique identifier of the account associated with the transaction
	//	@example	880e8400-e29b-41d4-a716-446655440000
	AccountId uuid.UUID `json:"accountId" binding:"required"`
	// Type is the type of the transaction (e.g., credit or debit) as an integer
	//	@example	0
	Type int `json:"type" binding:"required"`
	// Amount is the amount of the transaction
	//	@example	500
	Amount int64 `json:"amount" binding:"required"`
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
func (s *Server) handleGetTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionId, err := uuid.Parse(r.PathValue("transactionId"))
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusBadRequest, "Invalid transaction ID format", err)
			return
		}

		transaction, err := s.svc.Transactions.GetTransaction(r.Context(), transactionId)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusNotFound, "Transaction not found", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusOK, toTransactionResponse(transaction))
	}
}

// @Summary		Create a new transaction
// @Description	Creates a new transaction for an account
// @Produce		json
// @Param			transaction	body		TransactionPayload	true	"Transaction payload"
// @Success		201			{object}	TransactionResponse	"Transaction object"
// @Failure		400			{object}	ErrorResponse		"Invalid request payload"
// @Failure		500			{object}	ErrorResponse		"Failed to create transaction"
// @Router			/transactions [post]
func (s *Server) handleCreateTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p TransactionPayload

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // limit 1MB
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		err := dec.Decode(&p)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusBadRequest, "Invalid JSON payload", err)
			return
		}

		transaction, err := s.svc.Transactions.CreateTransaction(r.Context(), p.AccountId, p.Type, p.Amount)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusInternalServerError, "Failed to create transaction", err)
			return
		}

		err = s.svc.Accounts.UpdateAccountBalance(r.Context(), p.AccountId, p.Amount)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusInternalServerError, "Failed to update account balance", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusCreated, toTransactionResponse(transaction))
	}
}
