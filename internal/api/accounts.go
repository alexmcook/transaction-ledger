package api

import (
	"encoding/json"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// AccountResponse represents the account data returned in API responses
type AccountResponse struct {
	// Id is the unique identifier of the account
	//	@example	550e8400-e29b-41d4-a716-446655440000
	Id uuid.UUID `json:"id"`
	// UserId is the unique identifier of the user who owns the account
	//	@example	660e8400-e29b-41d4-a716-446655440000
	UserId uuid.UUID `json:"userId"`
	// Balance is the current balance of the account
	//	@example	1000
	Balance int64 `json:"balance"`
	// CreatedAt is the timestamp when the user was created
	//	@example	2025-12-25T11:11:00Z
	CreatedAt time.Time `json:"createdAt"`
}

// AccountPayload represents the account data received in API requests
type AccountPayload struct {
	// UserId is the unique identifier of the user who owns the account
	//	@example	660e8400-e29b-41d4-a716-446655440000
	UserId uuid.UUID `json:"userId" binding:"required"`
	// Balance is the initial balance of the account
	//	@example	1000
	Balance int64 `json:"balance" binding:"required"`
}

func toAccountResponse(a *model.Account) *AccountResponse {
	return &AccountResponse{
		Id:        a.Id,
		Balance:   a.Balance,
		CreatedAt: time.UnixMilli(a.CreatedAt),
	}
}

// @Summary		Get account
// @Description	Retrieves an account by its ID
// @Produce		json
// @Param			accountId	path		string			true	"Account ID"	format(uuid)
// @Success		200			{object}	model.Account	"Account object"
// @Failure		400			{object}	ErrorResponse	"Invalid account ID"
// @Failure		404			{object}	ErrorResponse	"Account not found"
// @Router			/accounts/{accountId} [get]
func (s *Server) handleGetAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountId, err := uuid.Parse(r.PathValue("accountId"))
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusBadRequest, "Invalid account ID format", err)
			return
		}

		account, err := s.svc.Accounts.GetAccount(r.Context(), accountId)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusNotFound, "Account not found", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusOK, toAccountResponse(account))
	}
}

// @Summary		Create a new account
// @Description	Creates a new account
// @Produce		json
// @Param			payload	body		AccountPayload	true	"Account creation payload"
// @Success		201		{object}	model.Account	"Account object"
// @Failure		400		{object}	ErrorResponse	"Invalid request payload"
// @Failure		500		{object}	ErrorResponse	"Failed to create account"
// @Router			/accounts [post]
func (s *Server) handleCreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p AccountPayload

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // limit 1MB
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		err := dec.Decode(&p)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusBadRequest, "Invalid JSON payload", err)
			return
		}

		account, err := s.svc.Accounts.CreateAccount(r.Context(), p.UserId, p.Balance)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusInternalServerError, "Failed to create account", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusCreated, toAccountResponse(account))
	}
}
