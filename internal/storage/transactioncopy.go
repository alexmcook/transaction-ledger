package storage

import (
	"github.com/alexmcook/transaction-ledger/internal/model"
)

// TransactionCopySource implements pgx.CopyFromSource
type TransactionCopySource struct {
	rows         []model.Transaction
	pos          int
	buf          []any
	partitionKey int16
}

func (s *TransactionCopySource) Next() bool {
	s.pos++
	return s.pos <= len(s.rows)
}

func (s *TransactionCopySource) Values() ([]any, error) {
	tx := s.rows[s.pos-1]
	s.buf[0] = tx.ID
	s.buf[1] = tx.AccountID
	s.buf[2] = tx.Amount
	s.buf[3] = tx.TransactionType
	s.buf[4] = tx.CreatedAt
	s.buf[5] = s.partitionKey
	return s.buf, nil
}

func (s *TransactionCopySource) Err() error {
	return nil
}
