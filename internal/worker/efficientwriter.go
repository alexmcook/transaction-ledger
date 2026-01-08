package worker

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"sync"
// 	"time"
//
// 	"github.com/alexmcook/transaction-ledger/internal/storage"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/twmb/franz-go/pkg/kgo"
// )
//
// type EfficientWriter struct {
// 	log        *slog.Logger
// 	db         *storage.PostgresStore
// 	client     *kgo.Client
// 	workChan   chan kgo.Fetches
// 	workerWg   sync.WaitGroup
// 	bufA       storage.EfficientTransactionSource
// 	bufB       storage.EfficientTransactionSource
// 	currentBuf *storage.EfficientTransactionSource
// }
//
// func NewEfficientWriter(log *slog.Logger, pool *pgxpool.Pool, client *kgo.Client) *EfficientWriter {
// 	return &EfficientWriter{
// 		log:    log,
// 		db:     storage.NewPostgresStore(log, pool),
// 		client: client,
// 	}
// }
//
// func (w *EfficientWriter) Start(ctx context.Context) error {
// 	w.bufA = *storage.NewEfficientTransactionSource()
// 	w.bufB = *storage.NewEfficientTransactionSource()
//
// 	w.workChan = make(chan kgo.Fetches, 2)
//
// 	w.workerWg.Add(1)
// 	// Clean context to allow graceful shutdown
// 	workerCtx := context.WithoutCancel(context.Background())
// 	go w.startWorker(workerCtx, w.workChan)
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		default:
// 			fetches := w.client.PollFetches(ctx)
// 			if fetches.IsClientClosed() {
// 				w.log.WarnContext(ctx, "Kafka client closed", slog.Any("ctx_err", ctx.Err()), slog.Any("client_err", w.client.Context().Err()))
// 				// TODO: retry reconnect
// 				return ctx.Err()
// 			}
// 			if fetches.Empty() {
// 				continue
// 			}
// 			w.workChan <- fetches
// 		}
// 	}
// }
//
// func (w *EfficientWriter) Stop(ctx context.Context) error {
// 	close(w.workChan)
//
// 	done := make(chan struct{})
// 	go func() {
// 		w.workerWg.Wait()
// 		close(done)
// 	}()
//
// 	select {
// 	case <-done:
// 		w.log.InfoContext(ctx, "All workers shut down gracefully")
// 		return nil
// 	case <-ctx.Done():
// 		return fmt.Errorf("shutdown timed out")
// 	}
// }
//
// func (w *EfficientWriter) swap(currentBuf *storage.EfficientTransactionSource) *storage.EfficientTransactionSource {
// 	if currentBuf == &w.bufA {
// 		return &w.bufB
// 	}
// 	return &w.bufA
// }
//
// func (w *EfficientWriter) startWorker(ctx context.Context, workChan chan kgo.Fetches) {
// 	currentBuf := &w.bufA
// 	defer w.workerWg.Done()
// 	var writeWg sync.WaitGroup
// 	for fetches := range workChan {
// 		iter := fetches.RecordIter()
//
// 		for !iter.Done() {
// 			currentBuf.Timestamp = time.Now()
//
// 			count := 0
// 			for !iter.Done() && count < 50000 {
// 				record := iter.Next()
// 				currentBuf.Txs[count].UnmarshalVT(record.Value)
// 				currentBuf.Offsets[record.Partition] = record.Offset
// 				count++
// 			}
// 			currentBuf.Count = count
//
// 			w.log.DebugContext(ctx, "Staging batch", slog.Int("count", currentBuf.Count))
//
// 			writeWg.Wait()
// 			writeWg.Add(1)
// 			go func(buf *storage.EfficientTransactionSource) {
// 				defer writeWg.Done()
// 				for {
// 					startBatch := time.Now()
// 					if err := w.db.Transactions().EfficientWriteBatch(ctx, buf); err != nil {
// 						w.log.ErrorContext(ctx, "Failed to write batch", slog.Int("count", buf.Count), slog.Any("error", err))
// 						time.Sleep(5 * time.Second)
// 						continue
// 					}
// 					dbWriteLatency.Observe(time.Since(startBatch).Seconds())
// 					transactionsStaged.Add(float64(buf.Count))
// 					break
// 				}
// 				buf.Reset() // Use local variable to avoid race condition
// 			}(currentBuf)
//
// 			currentBuf = w.swap(currentBuf)
// 		}
// 	}
// 	writeWg.Wait()
// }
