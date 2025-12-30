package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
)

type Store struct {
	logger       *slog.Logger
	Users        *UsersRepo
	Accounts     *AccountsRepo
	Transactions *TransactionsRepo
}

func Connect(ctx context.Context, maxConns int32) (*pgxpool.Pool, error) {
	dbUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	config.MaxConns = maxConns
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func NewStore(pool *pgxpool.Pool, logger *slog.Logger) *Store {
	return &Store{
		logger:       logger,
		Users:        NewUsersRepo(pool),
		Accounts:     NewAccountsRepo(pool),
		Transactions: NewTransactionsRepo(pool),
	}
}
