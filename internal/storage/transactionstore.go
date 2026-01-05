package storage

import (
	"context"
	"errors"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

type TransactionStore struct {
	pool *pgxpool.Pool
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

func (ts *TransactionStore) WriteBatch(ctx context.Context, batch []*kgo.Record) error {
	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	source := NewTransactionSource(batch)

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"staging"}, []string{"id", "account_id", "amount", "created_at"}, source)
	if err != nil {
		return err
	}

	const mergeStaging = `
		INSERT INTO transactions (id, account_id, amount, created_at)
		SELECT id, account_id, amount, created_at FROM staging
		ON CONFLICT (id) DO NOTHING
	`
	_, err = tx.Exec(ctx, mergeStaging)
	if err != nil {
		return err
	}

	offsets := source.Offsets()
	now := time.Now()
	batchUpdate := &pgx.Batch{}
	const kafkaOffset = `UPDATE kafka_offsets SET last_offset = $1, updated_at = $2 WHERE partition_id = $3`
	for partitionID, lastOffset := range offsets {
		batchUpdate.Queue(kafkaOffset, lastOffset, now, partitionID)
	}
	br := tx.SendBatch(ctx, batchUpdate)
	if err := br.Close(); err != nil {
		return err
	}

	const clearStaging = `TRUNCATE staging`
	_, err = tx.Exec(ctx, clearStaging)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
