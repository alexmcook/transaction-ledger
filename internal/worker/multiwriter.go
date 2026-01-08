package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/storage"
)

type MultiWriter struct {
	id         int
	log        *slog.Logger
	db         *storage.PostgresStore
	WorkChan   chan *RecordBatch
	workerWg   sync.WaitGroup
	bufA       *storage.EfficientTransactionSource
	bufB       *storage.EfficientTransactionSource
	currentBuf *storage.EfficientTransactionSource
}

func NewMultiWriter(id int, log *slog.Logger, db *storage.PostgresStore) *MultiWriter {
	return &MultiWriter{
		id:       id,
		log:      log,
		WorkChan: make(chan *RecordBatch, 4),
		db:       db,
	}
}

func (w *MultiWriter) Start(ctx context.Context) {
	w.bufA = storage.NewEfficientTransactionSource()
	w.bufB = storage.NewEfficientTransactionSource()

	w.workerWg.Add(1)
	go w.startWorker(ctx)
}

func (w *MultiWriter) Stop(ctx context.Context) error {
	close(w.WorkChan)

	done := make(chan struct{})
	go func() {
		w.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.log.InfoContext(ctx, "Worker stopped", slog.Int("worker_id", w.id))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out")
	}
}

func (w *MultiWriter) swap(currentBuf *storage.EfficientTransactionSource) *storage.EfficientTransactionSource {
	if currentBuf == w.bufA {
		return w.bufB
	}
	return w.bufA
}

func (w *MultiWriter) startWorker(ctx context.Context) {
	workerIDStr := strconv.Itoa(w.id)
	currentBuf := w.bufA
	defer w.workerWg.Done()
	var writeWg sync.WaitGroup
	for f := range w.WorkChan {
		currentBuf.Timestamp = time.Now()

		batch := f.Slab
		for i := range f.Count {
			currentBuf.Txs[i].UnmarshalVT(f.Slab[i].Value)
		}
		currentBuf.Offset = batch[f.Count-1].Offset
		currentBuf.Count = f.Count

		f.Reset()
		recordsPool.Put(f)

		w.log.DebugContext(ctx, "Staging batch", slog.Int("count", currentBuf.Count), slog.Int("worker_id", w.id))

		writeWg.Wait()
		writeWg.Add(1)
		go func(buf *storage.EfficientTransactionSource) {
			defer writeWg.Done()
			for {
				startBatch := time.Now()
				if err := w.db.Transactions().EfficientWriteBatch(ctx, w.id, buf); err != nil {
					w.log.ErrorContext(ctx, "Failed to write batch", slog.Int("count", buf.Count), slog.Any("error", err), slog.Int("worker_id", w.id))
					time.Sleep(5 * time.Second)
					select {
					case <-ctx.Done():
						return
					default:
					}
					continue
				}

				kafkaCommittedOffset.WithLabelValues(workerIDStr).Set(float64(buf.Offset))
				dbWriteLatency.Observe(time.Since(startBatch).Seconds())
				transactionsStaged.Add(float64(buf.Count))
				break
			}
			buf.Reset() // Use local variable to avoid race condition
		}(currentBuf)

		currentBuf = w.swap(currentBuf)
	}

	writeWg.Wait() // Ensure all writes are done before exiting
}
