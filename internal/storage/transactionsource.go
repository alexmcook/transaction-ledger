package storage

import (
	"encoding/binary"
	"sync/atomic"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

type TransactionSource struct {
	records []*kgo.Record
	idx     int
	pb      pb.Transaction
	offsets map[int32]int64
	salt    uint32
}

func NewTransactionSource(records []*kgo.Record) *TransactionSource {
	return &TransactionSource{
		records: records,
		idx:     -1,
		offsets: make(map[int32]int64),
		salt:    uint32(time.Now().UnixNano()),
	}
}

func (ts *TransactionSource) Next() bool {
	ts.idx++
	return ts.idx < len(ts.records)
}

func (ts *TransactionSource) Values() ([]any, error) {
	record := ts.records[ts.idx]
	err := proto.Unmarshal(record.Value, &ts.pb)
	if err != nil {
		return nil, err
	}

	ts.offsets[record.Partition] = record.Offset

	// LOAD TESTING modify ID to avoid collisions
	binary.BigEndian.PutUint32(ts.pb.Id[0:4], atomic.AddUint32(&ts.salt, 1))

	return []any{
		ts.pb.Id,
		ts.pb.AccountId,
		ts.pb.Amount,
		record.Timestamp,
	}, nil
}

func (ts *TransactionSource) Err() error {
	return nil
}

func (ts *TransactionSource) Offsets() map[int32]int64 {
	return ts.offsets
}
