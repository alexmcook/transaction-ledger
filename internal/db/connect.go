package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Users    *UsersRepo
	Accounts *AccountsRepo
}

func Connect(ctx context.Context, dbUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Users:    NewUsersRepo(pool),
		Accounts: NewAccountsRepo(pool),
	}
}
