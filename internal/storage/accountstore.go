package storage

import (
	"context"
	"errors"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountStore struct {
	pool *pgxpool.Pool
}

func (as *AccountStore) GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	const getAccountQuery = `SELECT id, user_id, balance, created_at FROM accounts WHERE id = $1`
	rows, err := as.pool.Query(ctx, getAccountQuery, id)
	if err != nil {
		return nil, err
	}

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Account])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Account not found
		}
		return nil, err
	}

	return &account, nil
}
