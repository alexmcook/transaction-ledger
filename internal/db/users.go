package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

func (r *UsersRepo) GetUser(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.pool.
		QueryRow(ctx, "SELECT id, created_at FROM accounts WHERE id = $1", id).
		Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UsersRepo) CreateUser(ctx context.Context) (*model.User, error) {
	var user model.User
	err := r.pool.
		QueryRow(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id, created_at").
		Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

