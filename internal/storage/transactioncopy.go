package storage

import (
	"encoding/binary"
	"time"

	pb "github.com/alexmcook/transaction-ledger/api/proto/v1"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// TransactionCopySource implements pgx.CopyFromSource
type TransactionCopySource struct {
	rows []api.CreateTransactionRequest
	pos  int

	// Define buffers for each column to avoid type conversion overhead
	idBuf   pgtype.UUID
	accBuf  pgtype.UUID
	amtBuf  pgtype.Int8
	typeBuf pgtype.Int2
	timeBuf pgtype.Timestamptz
	partBuf pgtype.Int2
	buf     []any

	now          time.Time
	rawTime      int64
	partitionKey int16

	baseUUID uuid.UUID
	seed     uint32
}

func (s *TransactionCopySource) Next() bool {
	s.pos++
	return s.pos <= len(s.rows)
}

func (s *TransactionCopySource) Values() ([]any, error) {
	tx := &s.rows[s.pos-1]

	// Generate a new UUID based on baseUUID and seed
	s.seed++
	binary.BigEndian.PutUint32(s.baseUUID[12:], s.seed)
	uid := s.baseUUID // snapshot of current baseUUID

	s.idBuf.Bytes = uid
	s.accBuf.Bytes = tx.AccountID
	s.amtBuf.Int64 = tx.Amount
	s.typeBuf.Int16 = int16(tx.TransactionType)
	s.timeBuf.Time = s.now
	s.partBuf.Int16 = s.partitionKey

	return s.buf, nil
}

func (s *TransactionCopySource) Err() error {
	return nil
}

func (s *TransactionCopySource) EncodeRowBinary(buf []byte, tx *api.CreateTransactionRequest) []byte {
	// Generate a new UUID based on baseUUID and seed
	s.seed++
	binary.BigEndian.PutUint32(s.baseUUID[12:], s.seed)

	// Number of columns
	buf = binary.BigEndian.AppendUint16(buf, 6)

	// Column 1: id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, s.baseUUID[:]...)

	// Column 2: account_id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, tx.AccountID[:]...)

	// Column 3: amount (INT8)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, uint64(tx.Amount))

	// Column 4: transaction_type (INT2)
	buf = binary.BigEndian.AppendUint32(buf, 2)
	buf = binary.BigEndian.AppendUint16(buf, uint16(tx.TransactionType))

	// Column 5: created_at (TIMESTAMPTZ)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, uint64(s.rawTime))

	// Column 6: partition_key (INT2)
	buf = binary.BigEndian.AppendUint32(buf, 2)
	buf = binary.BigEndian.AppendUint16(buf, uint16(s.partitionKey))

	return buf
}

func (s *TransactionCopySource) EncodeRowProtoBinary(buf []byte, tx *pb.CreateTransactionRequest) []byte {
	// Generate a new UUID based on baseUUID and seed
	s.seed++
	binary.BigEndian.PutUint32(s.baseUUID[12:], s.seed)

	// Number of columns
	buf = binary.BigEndian.AppendUint16(buf, 6)

	// Column 1: id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, s.baseUUID[:]...)

	// Column 2: account_id (UUID)
	buf = binary.BigEndian.AppendUint32(buf, 16)
	buf = append(buf, tx.AccountId[:]...)

	// Column 3: amount (INT8)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, uint64(tx.Amount))

	// Column 4: transaction_type (INT2)
	buf = binary.BigEndian.AppendUint32(buf, 2)
	buf = binary.BigEndian.AppendUint16(buf, uint16(tx.TransactionType))

	// Column 5: created_at (TIMESTAMPTZ)
	buf = binary.BigEndian.AppendUint32(buf, 8)
	buf = binary.BigEndian.AppendUint64(buf, uint64(s.rawTime))

	// Column 6: partition_key (INT2)
	buf = binary.BigEndian.AppendUint32(buf, 2)
	buf = binary.BigEndian.AppendUint16(buf, uint16(s.partitionKey))

	return buf
}
