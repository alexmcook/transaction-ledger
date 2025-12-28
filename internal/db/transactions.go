package db

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type TransactionsRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionsRepo(pool *pgxpool.Pool) *TransactionsRepo {
	return &TransactionsRepo{pool: pool}
}

func (r *TransactionsRepo) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	var transaction model.Transaction

	err := r.pool.
		QueryRow(ctx, "SELECT id, account_id, amount, transaction_type, created_at FROM transactions WHERE id = $1", id).
		Scan(&transaction.Id, &transaction.AccountId, &transaction.Amount, &transaction.Type, &transaction.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *TransactionsRepo) CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64) (*model.Transaction, error) {
	var transaction model.Transaction

	transactionId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	transaction.Id = transactionId
	transaction.AccountId = accountId
	transaction.Amount = amount
	transaction.Type = transactionType
	transaction.CreatedAt = time.Now().UnixMilli()

	_, err = r.pool.Exec(ctx, "INSERT INTO transactions (id, account_id, amount, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5)", transaction.Id, transaction.AccountId, transaction.Amount, transaction.Type, transaction.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}
