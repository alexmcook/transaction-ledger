package storage

import (
	"encoding/binary"
	"sync/atomic"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// LOAD TESTING salt to avoid ID collisions
var salt = uint32(time.Now().UnixNano())

type TransactionSource struct {
	records []*kgo.Record
	idx     int
	pb      pb.Transaction
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

	createdAt := time.Unix(0, ts.pb.CreatedAt)

	// LOAD TESTING modify ID to avoid collisions
	binary.BigEndian.PutUint32(ts.pb.Id[0:4], atomic.AddUint32(&salt, 1))

	return []any{
		ts.pb.Id,
		ts.pb.AccountId,
		ts.pb.Amount,
		createdAt,
	}, nil
}

func (ts *TransactionSource) Err() error {
	return nil
}
