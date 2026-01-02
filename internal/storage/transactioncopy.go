package storage

import (
	"time"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/google/uuid"
)

// TransactionCopySource implements pgx.CopyFromSource
type TransactionCopySource struct {
	rows         []api.CreateTransactionRequest
	pos          int
	buf          []any
	now          time.Time
	partitionKey int16
}

func (s *TransactionCopySource) Next() bool {
	s.pos++
	return s.pos <= len(s.rows)
}

func (s *TransactionCopySource) Values() ([]any, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	tx := s.rows[s.pos-1]
	s.buf[0] = uid
	s.buf[1] = tx.AccountID
	s.buf[2] = tx.Amount
	s.buf[3] = tx.TransactionType
	s.buf[4] = s.now
	s.buf[5] = s.partitionKey
	return s.buf, nil
}

func (s *TransactionCopySource) Err() error {
	return nil
}
