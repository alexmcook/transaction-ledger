package storage

import (
	"encoding/binary"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/jackc/pgx/v5/pgtype"
)

type EfficientTransactionSource struct {
	Txs       []pb.Transaction
	idx       int
	Count     int
	Offset    int64
	Timestamp time.Time

	idBuf  pgtype.UUID
	accBuf pgtype.UUID
	amtBuf pgtype.Int8
	tsBuf  pgtype.Timestamptz

	buf []any

	salt uint32
}

func NewEfficientTransactionSource() *EfficientTransactionSource {
	return &EfficientTransactionSource{
		Txs:  make([]pb.Transaction, 50000),
		idx:  -1,
		buf:  make([]any, 4),
		salt: uint32(time.Now().UnixNano()),
	}
}

func (ts *EfficientTransactionSource) Next() bool {
	ts.idx++
	return ts.idx < ts.Count
}

func (ts *EfficientTransactionSource) Values() ([]any, error) {
	tx := &ts.Txs[ts.idx]
	// LOAD TESTING modify ID to avoid collisions
	binary.BigEndian.PutUint32(tx.Id[0:4], ts.salt)
	ts.salt++

	copy(ts.idBuf.Bytes[:], tx.Id)
	ts.idBuf.Valid = true
	copy(ts.accBuf.Bytes[:], tx.AccountId)
	ts.accBuf.Valid = true

	ts.buf[0] = ts.idBuf
	ts.buf[1] = ts.accBuf
	ts.buf[2] = tx.Amount
	ts.buf[3] = ts.Timestamp

	return ts.buf, nil
}

func (ts *EfficientTransactionSource) Err() error {
	return nil
}

func (ts *EfficientTransactionSource) Reset() {
	ts.idx = -1
	ts.Offset = -1
}

func (ts *EfficientTransactionSource) EncodeRow(buf []byte, idx int, now uint64) []byte {
	tx := &ts.Txs[idx]
	ts.salt++
	binary.BigEndian.PutUint32(tx.Id[0:4], ts.salt)

	// Number of columns
	buf = binary.BigEndian.AppendUint16(buf, 4)

	// Column 1: id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, tx.Id[:]...)

	// Column 2: account_id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, tx.AccountId[:]...)

	// Column 3: amount (INT8)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, uint64(tx.Amount))

	// Column 4: created_at (TIMESTAMPTZ)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, now)

	return buf
}
