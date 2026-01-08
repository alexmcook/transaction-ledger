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
	Offsets   map[int32]int64
	Timestamp time.Time

	idBuf      pgtype.UUID
	accountBuf pgtype.UUID

	valuesBuf []any

	salt uint32
}

func NewEfficientTransactionSource() *EfficientTransactionSource {
	return &EfficientTransactionSource{
		Txs:       make([]pb.Transaction, 50000),
		idx:       -1,
		Offsets:   make(map[int32]int64),
		valuesBuf: make([]any, 4),
		salt:      uint32(time.Now().UnixNano()),
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
	copy(ts.accountBuf.Bytes[:], tx.AccountId)
	ts.accountBuf.Valid = true

	ts.valuesBuf[0] = ts.idBuf
	ts.valuesBuf[1] = ts.accountBuf
	ts.valuesBuf[2] = tx.Amount
	ts.valuesBuf[3] = ts.Timestamp

	return ts.valuesBuf, nil
}

func (ts *EfficientTransactionSource) Err() error {
	return nil
}

func (ts *EfficientTransactionSource) Reset() {
	ts.idx = -1
	for k := range ts.Offsets {
		delete(ts.Offsets, k)
	}
}
