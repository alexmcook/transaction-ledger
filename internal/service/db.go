package service

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type UserStore interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateUser(ctx context.Context) (*model.User, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error)
	CreateAccount(ctx context.Context, userId uuid.UUID, initialBalance int64) (*model.Account, error)
	UpdateAccountBalance(ctx context.Context, accountId uuid.UUID, amount int64) error
}

type TransactionStore interface {
	GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64, bucketId int32) (*model.Transaction, error)
}

type Service struct {
	logger         *slog.Logger
	Users          UserStore
	Accounts       AccountStore
	Transactions   TransactionStore
	BucketProvider db.BucketProvider
	TxChans        []chan model.TransactionPayload
	pool           *pgxpool.Pool
}

type Deps struct {
	Logger         *slog.Logger
	Users          UserStore
	Accounts       AccountStore
	Transactions   TransactionStore
	BucketProvider db.BucketProvider
	TxChans        []chan model.TransactionPayload
	Pool           *pgxpool.Pool
}

func New(d Deps) *Service {
	return &Service{
		logger:         d.Logger,
		Users:          d.Users,
		Accounts:       d.Accounts,
		Transactions:   d.Transactions,
		BucketProvider: d.BucketProvider,
		TxChans:       d.TxChans,
	}
}

func (s *Service) Shutdown() {
	for _, ch := range s.TxChans {
		close(ch)
	}

	time.Sleep(1 * time.Second) // Allow time for batch workers to finish

	s.pool.Close()
}
