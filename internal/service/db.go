package service

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"log/slog"
)

type UserStore interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateUser(ctx context.Context) (*model.User, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error)
	CreateAccount(ctx context.Context, userId uuid.UUID, initialBalance int64) (*model.Account, error)
}

type TransactionStore interface {
	GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64) (*model.Transaction, error)
}

type Service struct {
	logger       *slog.Logger
	Users        UserStore
	Accounts     AccountStore
	Transactions TransactionStore
}

type Deps struct {
	Logger       *slog.Logger
	Users        UserStore
	Accounts     AccountStore
	Transactions TransactionStore
}

func New(d Deps) *Service {
	return &Service{
		logger:       d.Logger,
		Users:        d.Users,
		Accounts:     d.Accounts,
		Transactions: d.Transactions,
	}
}
