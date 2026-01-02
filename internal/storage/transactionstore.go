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

type TransactionStore struct {
	pool             *pgxpool.Pool
	partitionProvder PartitionProvider
}

func (ts *TransactionStore) CreateTransaction(ctx context.Context, params api.CreateTransactionParams) error {
	const createTransactionQuery = `INSERT INTO transactions (id, account_id, amount, transaction_type, created_at, partition_key) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := ts.pool.Exec(ctx, createTransactionQuery, params.ID, params.AccountID, params.Amount, params.TransactionType, params.CreatedAt, ts.partitionProvder.GetActivePartition())
	return err
}

func (ts *TransactionStore) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	const getTransactionQuery = `SELECT id, account_id, amount, transaction_type, created_at FROM transactions WHERE id = $1`
	rows, err := ts.pool.Query(ctx, getTransactionQuery, id)
	if err != nil {
		return nil, err
	}

	tx, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Transaction])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Transaction not found
		}
		return nil, err
	}

	return &tx, nil
}

func (ts *TransactionStore) CreateBatchTransaction(ctx context.Context, txs []model.Transaction) (int64, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	source := &TransactionCopySource{
		rows:         txs,
		pos:          0,
		buf:          make([]any, 6),
		partitionKey: ts.partitionProvder.GetActivePartition(),
	}

	count, err := ts.pool.CopyFrom(
		ctx,
		pgx.Identifier{"transactions"},
		[]string{"id", "account_id", "amount", "transaction_type", "created_at", "bucket_id"},
		source,
	)

	return count, err
}
