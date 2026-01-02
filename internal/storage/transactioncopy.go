package storage

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/google/uuid"
)

// TransactionCopySource implements pgx.CopyFromSource
type TransactionCopySource struct {
	rows []api.CreateTransactionRequest
	pos  int
	buf  []any

	now          time.Time
	partitionKey int16

	baseUUID uuid.UUID
	seed     uint32
}

func NewTransactionCopySource(txs []api.CreateTransactionRequest, partitionKey int16) *TransactionCopySource {
	uid, _ := uuid.NewV7()

	return &TransactionCopySource{
		rows:         txs,
		pos:          0,
		buf:          make([]any, 6),
		now:          time.Now(),
		partitionKey: partitionKey,
		baseUUID:     uid,
		seed:         rand.Uint32(),
	}
}

func (s *TransactionCopySource) Next() bool {
	s.pos++
	return s.pos <= len(s.rows)
}

func (s *TransactionCopySource) Values() ([]any, error) {
	// Generate a new UUID based on baseUUID and seed
	s.seed++
	binary.BigEndian.PutUint32(s.baseUUID[12:], s.seed)
	uid := s.baseUUID // snapshot of current baseUUID

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
