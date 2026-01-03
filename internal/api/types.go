package api

import (
	"context"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type AccountResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type SingleTransactionResponse struct {
	CreatedCount int `json:"created_count"`
}

type BatchTransactionResponse struct {
	CreatedCount int `json:"created_count"`
}

type StoreRegistry interface {
	Accounts() AccountStore
	Transactions() TransactionStore
}

type AccountStore interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error)
}

type TransactionStore interface {
	GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
}
