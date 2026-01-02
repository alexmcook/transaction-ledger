package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (ts *TransactionStore) CreateTransaction(ctx context.Context, tx api.CreateTransactionRequest) error {
	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	activePartition := ts.partitionProvder.GetActivePartition()
	createTransactionQuery := fmt.Sprintf(`INSERT INTO transactions_p%d (id, account_id, amount, transaction_type, created_at, partition_key) VALUES ($1, $2, $3, $4, $5, $6)`, activePartition)
	_, err = ts.pool.Exec(ctx, createTransactionQuery, uid, tx.AccountID, tx.Amount, tx.TransactionType, time.Now(), activePartition)
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

func (ts *TransactionStore) CreateBatchTransaction(ctx context.Context, txs []api.CreateTransactionRequest) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	activePartition := ts.partitionProvder.GetActivePartition()
	source := NewTransactionCopySource(txs, activePartition)

	partitionStr := fmt.Sprintf("transactions_p%d", activePartition)
	count, err := ts.pool.CopyFrom(
		ctx,
		pgx.Identifier{partitionStr},
		[]string{"id", "account_id", "amount", "transaction_type", "created_at", "partition_key"},
		source,
	)

	return int(count), err
}
