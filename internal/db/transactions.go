package db

import (
	"context"
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
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

func (r *TransactionsRepo) CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64, bucketId int32) (*model.Transaction, error) {
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

	// Insert into the active bucket partition
	_, err = r.pool.Exec(ctx, "INSERT INTO transactions (id, account_id, amount, transaction_type, created_at, bucket_id) VALUES ($1, $2, $3, $4, $5, $6)", transaction.Id, transaction.AccountId, transaction.Amount, transaction.Type, transaction.CreatedAt, bucketId)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *TransactionsRepo) FlushBucket(ctx context.Context, bucketId int32) error {
	if bucketId != 0 && bucketId != 1 {
		return fmt.Errorf("invalid bucketId: %d", bucketId)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Swap partitions
	query := fmt.Sprintf("ALTER TABLE transactions DETACH PARTITION tx_buf_%d", bucketId)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Bulk update account balances
	// ORDER BY account_id to avoid deadlocks
	query = fmt.Sprintf(`
		UPDATE accounts a
		SET balance = a.balance + u.sum_amount
		FROM (
			SELECT account_id, SUM(amount) as sum_amount
			FROM tx_buf_%d
			GROUP BY account_id
			ORDER BY account_id
		) AS u
		WHERE a.id = u.account_id
	`, bucketId)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Truncate the detached partition
	query = fmt.Sprintf("TRUNCATE TABLE tx_buf_%d", bucketId)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Re-attach the detached partition as the new buffer
	query = fmt.Sprintf("ALTER TABLE transactions ATTACH PARTITION tx_buf_%d FOR VALUES IN (%d)", bucketId, bucketId)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

func (r *TransactionsRepo) BatchProcess(ctx context.Context, txs []*model.Transaction, bucketId int32) error {
	if bucketId != 0 && bucketId != 1 {
		return fmt.Errorf("invalid bucketId: %d", bucketId)
	}

	partition := fmt.Sprintf("tx_buf_%d", bucketId)

	rows, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{partition}, // Directly target the partition
		[]string{"id", "account_id", "amount", "transaction_type", "created_at", "bucket_id"},
		pgx.CopyFromSlice(len(txs), func(i int) ([]any, error) {
			return []any{
				txs[i].Id,
				txs[i].AccountId,
				txs[i].Amount,
				txs[i].Type,
				txs[i].CreatedAt,
				bucketId,
			}, nil
		}),
	)
	if err != nil {
		return err
	}

	if int(rows) != len(txs) {
		return fmt.Errorf("expected to insert %d rows, but inserted %d", len(txs), rows)
	}

	return nil
}
