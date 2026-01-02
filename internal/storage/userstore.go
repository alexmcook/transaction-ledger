package storage

import (
	"context"
	"errors"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func (us *UserStore) CreateUser(ctx context.Context, params api.CreateUserParams) error {
	const createUserQuery = `INSERT INTO users (id, created_at) VALUES ($1, $2)`
	_, err := us.pool.Exec(ctx, createUserQuery, params.ID, params.CreatedAt)
	return err
}

func (us *UserStore) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const getUserQuery = `SELECT id, created_at FROM users WHERE id = $1`
	rows, err := us.pool.Query(ctx, getUserQuery, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}
