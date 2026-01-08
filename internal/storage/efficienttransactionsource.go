package storage

import (
	"encoding/binary"
	"sync/atomic"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
)

type EfficientTransactionSource struct {
	Txs       []pb.Transaction
	idx       int
	Count     int
	Offsets   map[int32]int64
	Timestamp time.Time
	buf       []any
}

func NewEfficientTransactionSource() *EfficientTransactionSource {
	return &EfficientTransactionSource{
		Txs:     make([]pb.Transaction, 50000),
		idx:     -1,
		Offsets: make(map[int32]int64),
		buf:     make([]any, 4),
	}
}

func (ts *EfficientTransactionSource) Next() bool {
	ts.idx++
	return ts.idx < ts.Count
}

func (ts *EfficientTransactionSource) Values() ([]any, error) {
	// LOAD TESTING modify ID to avoid collisions
	binary.BigEndian.PutUint32(ts.Txs[ts.idx].Id[0:4], atomic.AddUint32(&salt, 1))

	ts.buf[0] = ts.Txs[ts.idx].Id
	ts.buf[1] = ts.Txs[ts.idx].AccountId
	ts.buf[2] = ts.Txs[ts.idx].Amount
	ts.buf[3] = ts.Timestamp

	return ts.buf, nil
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
