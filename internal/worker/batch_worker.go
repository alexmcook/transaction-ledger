package worker

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/db"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"log/slog"
	"sync"
	"time"
)

type Batchable interface {
	BatchProcess(ctx context.Context, batch []*model.Transaction, bucketId int32) error
}

type BatchWorker struct {
	logger         *slog.Logger
	txChan         chan *model.Transaction
	batchSize      int
	batchInterval  time.Duration
	batchable      Batchable
	bucketProvider db.BucketProvider
	workerOnce     sync.Once
}

type BatchWorkerOpts struct {
	Logger         *slog.Logger
	TxChan         chan *model.Transaction
	BatchSize      int
	BatchInterval  time.Duration
	Batchable      Batchable
	BucketProvider db.BucketProvider
}

func NewBatchWorker(opts BatchWorkerOpts) *BatchWorker {
	return &BatchWorker{
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
		bw.logger.Debug("BatchWorker: starting batch worker", slog.Duration("batch_interval", bw.batchInterval), slog.Int("batch_size", bw.batchSize))
		go bw.batchRoutine(ctx)
	})
}

func (bw *BatchWorker) batchRoutine(ctx context.Context) {
	batch := make([]*model.Transaction, 0, bw.batchSize)
	ticker := time.NewTicker(bw.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case tx, ok := <-bw.txChan:
			if !ok {
				bw.logger.Debug("BatchWorker: txChan closed")
				if len(batch) > 0 {
					bw.processBatch(batch)
				}
				return
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
			bw.logger.Debug("BatchWorker: shutting down")
			if len(batch) > 0 {
				bw.processBatch(batch)
			}
			return
		}
	}
}

func (bw *BatchWorker) processBatch(batch []*model.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bucketId := bw.bucketProvider.GetActiveBucket()

	bw.logger.Debug("BatchWorker: processing batch", slog.Int("batch_size", len(batch)), slog.Int("bucket", int(bucketId)))
	err := bw.batchable.BatchProcess(ctx, batch, bucketId)
	if err != nil {
		bw.logger.Error("BatchWorker: error processing batch", slog.String("error", err.Error()))
	}
}
