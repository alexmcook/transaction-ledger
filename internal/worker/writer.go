package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

type WriterInterface interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Writer struct {
	log      *slog.Logger
	db       *storage.PostgresStore
	client   *kgo.Client
	workChan chan []*kgo.Record
	workerWg sync.WaitGroup
}

func NewWriter(log *slog.Logger, pool *pgxpool.Pool, client *kgo.Client) *Writer {
	return &Writer{
		log:    log,
		db:     storage.NewPostgresStore(log, pool),
		client: client,
	}
}

func (w *Writer) Start(ctx context.Context) error {
	w.workChan = make(chan []*kgo.Record, 2)

	w.workerWg.Add(1)
	// Clean context to allow graceful shutdown
	workerCtx := context.WithoutCancel(context.Background())
	go w.startWorker(workerCtx, w.workChan)

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return w.Stop(shutdownCtx)
		default:
			fetchStart := time.Now()
			fetches := w.client.PollRecords(ctx, 50000)
			if fetches.IsClientClosed() {
				w.log.WarnContext(ctx, "Kafka client closed", slog.Any("ctx_err", ctx.Err()), slog.Any("client_err", w.client.Context().Err()))
				// TODO: retry reconnect
				return nil
			}
			fetchLatency.Observe(time.Since(fetchStart).Seconds())

			w.log.DebugContext(ctx, "Fetched records", slog.Int("count", len(fetches.Records())))

			if len(fetches.Records()) == 0 {
				continue
			}

			w.workChan <- fetches.Records()
		}
	}
}

func (w *Writer) Stop(ctx context.Context) error {
	close(w.workChan)

	done := make(chan struct{})
	go func() {
		w.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.log.InfoContext(ctx, "All workers shut down gracefully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out")
	}
}

func (w *Writer) startWorker(ctx context.Context, workChan chan []*kgo.Record) {
	defer w.workerWg.Done()
	for {
		select {
		case work, ok := <-workChan:
			if !ok {
				return
			}

			for {
				startBatch := time.Now()
				if err := w.db.Transactions().WriteBatch(ctx, work); err != nil {
					w.log.ErrorContext(ctx, "Failed to write batch", slog.Int("count", len(work)), slog.Any("error", err))
					time.Sleep(5 * time.Second)
					continue
				}
				dbWriteLatency.Observe(time.Since(startBatch).Seconds())
				transactionsStaged.Add(float64(len(work)))
				break
			}

			w.log.DebugContext(ctx, "Committing records", slog.Int("count", len(work)))
			w.client.CommitRecords(ctx, work...)

		case <-time.After(20 * time.Second):
			w.log.DebugContext(ctx, "Worker idle")
		}
	}
}
