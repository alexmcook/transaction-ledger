package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	pb "github.com/alexmcook/transaction-ledger/api/proto/v1"
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionStore struct {
	pool             *pgxpool.Pool
	partitionProvder PartitionProvider
}

var sourcePool = sync.Pool{
	New: func() any {
		s := &TransactionCopySource{}

		s.idBuf.Valid = true
		s.accBuf.Valid = true
		s.amtBuf.Valid = true
		s.typeBuf.Valid = true
		s.timeBuf.Valid = true
		s.partBuf.Valid = true

		s.buf = []any{
			&s.idBuf,
			&s.accBuf,
			&s.amtBuf,
			&s.typeBuf,
			&s.timeBuf,
			&s.partBuf,
		}

		return s
	},
}

func (ts *TransactionStore) CreateTransaction(ctx context.Context, tx api.CreateTransactionRequest) error {
	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	activePartition := ts.partitionProvder.GetActivePartition()
	createTransactionQuery := fmt.Sprintf(`INSERT INTO transactions_p%d (id, account_id, amount, transaction_type, created_at, partition_key) VALUES ($1, $2, $3, $4, $5, $6)`, activePartition)
	_, err = ts.pool.Exec(ctx, createTransactionQuery, uid, tx.AccountID, tx.Amount, tx.TransactionType, time.Now(), activePartition)
	return err
}

func (ts *TransactionStore) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	const getTransactionQuery = `SELECT id, account_id, amount, transaction_type, created_at FROM transactions WHERE id = $1`
	rows, err := ts.pool.Query(ctx, getTransactionQuery, id)
	if err != nil {
		return nil, err
	}

	tx, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Transaction])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Transaction not found
		}
		return nil, err
	}

	return &tx, nil
}

func (ts *TransactionStore) CreateBatchTransaction(ctx context.Context, txs []api.CreateTransactionRequest) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	activePartition := ts.partitionProvder.GetActivePartition()
	source := sourcePool.Get().(*TransactionCopySource)
	source.rows = txs
	source.pos = 0
	source.now = time.Now()
	source.partitionKey = activePartition
	uid, _ := uuid.NewV7()
	source.baseUUID = uid
	source.seed = rand.Uint32()

	defer func() {
		source.rows = nil
		sourcePool.Put(source)
	}()

	partitionStr := fmt.Sprintf("transactions_p%d", activePartition)
	count, err := ts.pool.CopyFrom(
		ctx,
		pgx.Identifier{partitionStr},
		[]string{"id", "account_id", "amount", "transaction_type", "created_at", "partition_key"},
		source,
	)

	return int(count), err
}

func (ts *TransactionStore) CreateBinaryBatchTransaction(ctx context.Context, txs []api.CreateTransactionRequest) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	buf := make([]byte, 0, 15+(len(txs)*54)+2) // Preallocate buffer
	buf = append(buf, "PGCOPY\n\xff\r\n\x00"...)
	buf = binary.BigEndian.AppendUint32(buf, 0) // Flags
	buf = binary.BigEndian.AppendUint32(buf, 0) // Header extension area size

	activePartition := ts.partitionProvder.GetActivePartition()
	source := sourcePool.Get().(*TransactionCopySource)

	now := time.Now()
	source.rawTime = (now.Unix()-946684800)*1e6 + int64(now.Nanosecond()/1e3) // Microseconds since 2000-01-01
	source.partitionKey = activePartition
	uid, _ := uuid.NewV7()
	source.baseUUID = uid
	source.seed = rand.Uint32()
	defer func() {
		sourcePool.Put(source)
	}()

	for i := range txs {
		buf = source.EncodeRowBinary(buf, &txs[i])
	}

	buf = binary.BigEndian.AppendUint16(buf, 0xffff) // End of copy marker

	conn, err := ts.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	rawConn := conn.Conn().PgConn()
	_, err = rawConn.CopyFrom(ctx, bytes.NewReader(buf), fmt.Sprintf("COPY transactions_p%d FROM STDIN WITH (FORMAT BINARY)", activePartition))

	return len(txs), err
}

func (ts *TransactionStore) CreateProtoBinaryBatchTransaction(ctx context.Context, batch *pb.TransactionBatch) (int, error) {
	txs := batch.GetTransactions()
	if len(txs) == 0 {
		return 0, nil
	}

	buf := make([]byte, 0, 15+(len(txs)*54)+2) // Preallocate buffer
	buf = append(buf, "PGCOPY\n\xff\r\n\x00"...)
	buf = binary.BigEndian.AppendUint32(buf, 0) // Flags
	buf = binary.BigEndian.AppendUint32(buf, 0) // Header extension area size

	activePartition := ts.partitionProvder.GetActivePartition()
	source := sourcePool.Get().(*TransactionCopySource)

	now := time.Now()
	source.rawTime = (now.Unix()-946684800)*1e6 + int64(now.Nanosecond()/1e3) // Microseconds since 2000-01-01
	source.partitionKey = activePartition
	uid, _ := uuid.NewV7()
	source.baseUUID = uid
	source.seed = rand.Uint32()
	defer func() {
		sourcePool.Put(source)
	}()

	for _, tx := range txs {
		buf = source.EncodeRowProtoBinary(buf, tx)
	}

	buf = binary.BigEndian.AppendUint16(buf, 0xffff) // End of copy marker

	conn, err := ts.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	rawConn := conn.Conn().PgConn()
	_, err = rawConn.CopyFrom(ctx, bytes.NewReader(buf), fmt.Sprintf("COPY transactions_p%d FROM STDIN WITH (FORMAT BINARY)", activePartition))

	return len(txs), err
}
