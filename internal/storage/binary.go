package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"time"

	"github.com/jackc/pgx/v5"
)

/*
** Efficient WriteBatch implementation, utilizing zero-copy techniques to direct binary copy into staging table
 */
func (ts *TransactionStore) EfficientWriteBatch(ctx context.Context, source *EfficientTransactionSource) error {
	now := time.Now()

	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	buf := make([]byte, 0, 15+(source.Count*39)+2) // Preallocate buffer
	buf = append(buf, "PGCOPY\n\xff\r\n\x00"...)
	buf = binary.BigEndian.AppendUint32(buf, 0) // Flags
	buf = binary.BigEndian.AppendUint32(buf, 0) // Header extension area size

	rawTime := uint64((now.Unix()-946684800)*1e6 + int64(now.Nanosecond()/1e3)) // Microseconds since 2000-01-01

	for i := 0; i < source.Count; i++ {
		buf = source.EncodeRow(buf, i, rawTime)
	}
	buf = binary.BigEndian.AppendUint16(buf, 0xffff) // End of copy marker

	const createTemp = `
		CREATE TEMP TABLE staging (
		  LIKE transactions
		) ON COMMIT DROP
		`
	_, err = tx.Exec(ctx, createTemp)
	if err != nil {
		return err
	}

	rawConn := tx.Conn().PgConn()
	_, err = rawConn.CopyFrom(ctx, bytes.NewReader(buf), "COPY staging FROM STDIN WITH (FORMAT BINARY)")
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
