package db

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

func (r *UsersRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.pool.
		QueryRow(ctx, "SELECT id, created_at FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UsersRepo) CreateUser(ctx context.Context) (*model.User, error) {
	var user model.User

	userId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	user.Id = userId

	time := time.Now().UnixMilli()
	user.CreatedAt = time

	_, err = r.pool.Exec(ctx, "INSERT INTO users (id, created_at) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING", user.Id, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

