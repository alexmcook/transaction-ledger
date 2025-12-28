package db

import (
	"context"
	"log/slog"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	logger  *slog.Logger
	Users    *UsersRepo
	Accounts *AccountsRepo
	Transactions *TransactionsRepo
}

func Connect(ctx context.Context, dbUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func NewStore(pool *pgxpool.Pool, logger *slog.Logger) *Store {
	return &Store{
		logger:	 logger,
		Users:    NewUsersRepo(pool),
		Accounts: NewAccountsRepo(pool),
		Transactions: NewTransactionsRepo(pool),
	}
}
