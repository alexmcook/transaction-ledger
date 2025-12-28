package db

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type AccountsRepo struct {
	pool *pgxpool.Pool
}

func NewAccountsRepo(pool *pgxpool.Pool) *AccountsRepo {
	return &AccountsRepo{pool: pool}
}

func (r *AccountsRepo) GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	var account model.Account

	err := r.pool.
		QueryRow(ctx, "SELECT id, user_id, balance, created_at FROM accounts WHERE id = $1", id).
		Scan(&account.Id, &account.UserId, &account.Balance, &account.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AccountsRepo) CreateAccount(ctx context.Context, userId uuid.UUID, initialBalance int64) (*model.Account, error) {
	var account model.Account
	accountId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	account.Id = accountId
	account.UserId = userId
	account.Balance = initialBalance
	account.CreatedAt = time.Now().UnixMilli()

	_, err = r.pool.Exec(ctx, "INSERT INTO accounts (id, user_id, balance, created_at) VALUES ($1, $2, $3)", account.Id, account.UserId, account.Balance, account.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &account, nil
}
