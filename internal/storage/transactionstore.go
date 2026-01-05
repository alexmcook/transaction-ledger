package storage

import (
	"context"
	"errors"
	"fmt"

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

func (ts *TransactionStore) WriteBatch(ctx context.Context, shardID int, batch []*kgo.Record) error {
	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	stagingTable := fmt.Sprintf("staging_%d", shardID)
	source := &TransactionSource{records: batch, idx: -1}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{stagingTable}, []string{"id", "account_id", "amount", "created_at"}, source)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO transactions (id, account_id, amount, created_at)
		SELECT id, account_id, amount, created_at FROM %s
		ON CONFLICT (id) DO NOTHING
	`, stagingTable))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s`, stagingTable))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
