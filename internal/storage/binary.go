package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

/*
** Efficient WriteBatch implementation, utilizing zero-copy techniques to direct binary copy into staging table
 */
func (ts *TransactionStore) EfficientWriteBatch(ctx context.Context, source *EfficientTransactionSource) error {
	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const createTemp = `
		CREATE TEMP TABLE staging (
		  LIKE transactions
		) ON COMMIT DROP
		`
	_, err = tx.Exec(ctx, createTemp)
	if err != nil {
		return err
	}

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

	now := time.Now()
	batchUpdate := &pgx.Batch{}
	const kafkaOffset = `UPDATE kafka_offsets SET last_offset = $1, updated_at = $2 WHERE partition_id = $3`
	for partitionID, lastOffset := range source.Offsets {
		batchUpdate.Queue(kafkaOffset, lastOffset, now, partitionID)
	}
	br := tx.SendBatch(ctx, batchUpdate)
	if err := br.Close(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
