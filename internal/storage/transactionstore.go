package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionStore struct {
	pool            *pgxpool.Pool
	copyQueries     [64]string
	mergeQueries    [64]string
	truncateQueries [64]string
}

func NewTransactionStore(pool *pgxpool.Pool) *TransactionStore {
	ts := &TransactionStore{
		pool: pool,
	}

	for i := range 64 {
		ts.copyQueries[i] = fmt.Sprintf(`COPY staging_%d FROM STDIN WITH (FORMAT BINARY)`, i)
		ts.mergeQueries[i] = fmt.Sprintf(`
		INSERT INTO transactions_%d (id, account_id, amount, created_at)
		SELECT id, account_id, amount, created_at FROM staging_%d
		ON CONFLICT (id) DO NOTHING
	`, i, i)
		ts.truncateQueries[i] = fmt.Sprintf(`TRUNCATE TABLE staging_%d`, i)
	}

	return ts
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
