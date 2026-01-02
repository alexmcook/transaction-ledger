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

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAccountRequest struct {
	UserID  uuid.UUID `json:"user_id"`
	Balance int64     `json:"balance"`
}

type AccountResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	Type      int8      `json:"type"`
	Amount    int64     `json:"amount"`
}

type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	Amount    int64     `json:"amount"`
	Type      int8      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreRegistry interface {
	Users() UserStore
	Accounts() AccountStore
	Transactions() TransactionStore
}

type UserStore interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateUser(ctx context.Context, params CreateUserParams) error
}

type AccountStore interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error)
	CreateAccount(ctx context.Context, params CreateAccountParams) error
}

type TransactionStore interface {
	GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, params CreateTransactionParams) error
}

type CreateUserParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type CreateAccountParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   int64
	CreatedAt time.Time
}

type CreateTransactionParams struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Amount    int64
	Type      int8
	CreatedAt time.Time
}
