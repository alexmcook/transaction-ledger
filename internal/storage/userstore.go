package storage

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func (us *UserStore) CreateUser(ctx context.Context, id uuid.UUID, createdAt time.Time) error {
	const createUserQuery = `INSERT INTO users id, created_at VALUES ($1, $2)`
	_, err := us.pool.Exec(ctx, createUserQuery, id, createdAt)
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
		return nil, err
	}

	return &user, nil
}
