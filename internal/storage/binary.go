package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"time"
)

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 15+(1000*39)+2) // Preallocate buffer for 1000 rows
		return &b
	},
}

/*
** Efficient WriteBatch implementation, utilizing zero-copy techniques to direct binary copy into staging table
 */
func (ts *TransactionStore) EfficientWriteBatch(ctx context.Context, workerId int, source *EfficientTransactionSource) error {
	now := time.Now()

	tx, err := ts.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear staging table at start of transaction to minimize locks
	_, err = tx.Exec(ctx, ts.truncateQueries[workerId])
	if err != nil {
		return err
	}

	bPtr := bufPool.Get().(*[]byte)
	buf := (*bPtr)[:0]
	defer func() {
		*bPtr = buf[:0]
		bufPool.Put(bPtr)
	}()

	buf = append(buf, "PGCOPY\n\xff\r\n\x00"...)
	buf = binary.BigEndian.AppendUint32(buf, 0) // Flags
	buf = binary.BigEndian.AppendUint32(buf, 0) // Header extension area size

	rawTime := uint64((now.Unix()-946684800)*1e6 + int64(now.Nanosecond()/1e3)) // Microseconds since 2000-01-01

	for i := 0; i < source.Count; i++ {
		buf = source.EncodeRow(buf, i, rawTime)
	}
	buf = binary.BigEndian.AppendUint16(buf, 0xffff) // End of copy marker

	rawConn := tx.Conn().PgConn()
	_, err = rawConn.CopyFrom(ctx, bytes.NewReader(buf), ts.copyQueries[workerId])
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, ts.mergeQueries[workerId])
	if err != nil {
		return err
	}

	const kafkaOffset = `UPDATE kafka_offsets SET last_offset = $1, updated_at = $2 WHERE partition_id = $3`
	_, err = tx.Exec(ctx, kafkaOffset, source.Offset, now, workerId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
