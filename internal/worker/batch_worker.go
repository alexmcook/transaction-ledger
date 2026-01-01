package worker

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"log/slog"
	"sync"
	"time"
)

type Batchable interface {
	BatchProcess(ctx context.Context, batch []model.Transaction, bucketId int32) error
}

type BatchWorker struct {
	id             int
	logger         *slog.Logger
	txChan         chan model.TransactionPayload
	batchSize      int
	batchInterval  time.Duration
	batchable      Batchable
	bucketProvider db.BucketProvider
	workerOnce     sync.Once
}

type BatchWorkerOpts struct {
	Logger         *slog.Logger
	TxChan         chan model.TransactionPayload
	BatchSize      int
	BatchInterval  time.Duration
	Batchable      Batchable
	BucketProvider db.BucketProvider
}

func NewBatchWorker(workerId int, opts BatchWorkerOpts) *BatchWorker {
	return &BatchWorker{
		id:             workerId,
		logger:         opts.Logger,
		txChan:         opts.TxChan,
		batchSize:      opts.BatchSize,
		batchInterval:  opts.BatchInterval,
		batchable:      opts.Batchable,
		bucketProvider: opts.BucketProvider,
	}
}

func (bw *BatchWorker) Start(ctx context.Context) {
	bw.workerOnce.Do(func() {
		bw.logger.Debug("BatchWorker: starting batch worker", slog.Int("id", bw.id), slog.Duration("batch_interval", bw.batchInterval), slog.Int("batch_size", bw.batchSize))
		go bw.batchRoutine(ctx)
	})
}

func (bw *BatchWorker) batchRoutine(ctx context.Context) {
	batch := make([]model.Transaction, 0, bw.batchSize)
	ticker := time.NewTicker(bw.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case p, ok := <-bw.txChan:
			if !ok {
				bw.logger.Debug("BatchWorker: txChan closed", slog.Int("id", bw.id))
				if len(batch) > 0 {
					bw.processBatch(batch)
				}
				return
			}

			tx := model.Transaction{
				Id:						 	 uuid.Must(uuid.NewV7()),
				AccountId:       p.AccountId,
				Amount:          p.Amount,
				Type:            p.Type,
				CreatedAt:       time.Now().UnixMilli(),
			}

			batch = append(batch, tx)
			if len(batch) >= bw.batchSize {
				bw.processBatch(batch)
				batch = batch[:0] // Reset batch
				ticker.Reset(bw.batchInterval)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				bw.processBatch(batch)
				batch = batch[:0] // Reset batch
			}
		case <-ctx.Done():
			bw.logger.Debug("BatchWorker: shutting down", slog.Int("id", bw.id))
			if len(batch) > 0 {
				bw.processBatch(batch)
			}
			return
		}
	}
}

func (bw *BatchWorker) processBatch(batch []model.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bucketId := bw.bucketProvider.GetActiveBucket()

	bw.logger.Debug("BatchWorker: processing batch", slog.Int("id", bw.id), slog.Int("batch_size", len(batch)), slog.Int("bucket", int(bucketId)))
	err := bw.batchable.BatchProcess(ctx, batch, bucketId)
	if err != nil {
		bw.logger.Error("BatchWorker: error processing batch", slog.Int("bw", bw.id), slog.String("error", err.Error()))
	}
}
