package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type TransactionsRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionsRepo(pool *pgxpool.Pool) *TransactionsRepo {
	return &TransactionsRepo{pool: pool}
}

func (r *TransactionsRepo) GetTransaction(ctx context.Context, id int64) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.pool.
		QueryRow(ctx, "SELECT id, account_id, amount, created_at FROM transactions WHERE id = $1", id).
		Scan(&transaction.Id, &transaction.AccountId, &transaction.Amount, &transaction.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionsRepo) CreateTransaction(ctx context.Context, accountId int64, amount int64) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.pool.
		QueryRow(ctx, "INSERT INTO transactions (account_id, amount) VALUES ($1, $2) RETURNING id, account_id, amount, created_at", accountId, amount).
		Scan(&transaction.Id, &transaction.AccountId, &transaction.Amount, &transaction.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}
