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

type Work struct {
	Records []*kgo.Record
	Done    chan error
}

type Writer struct {
	log      *slog.Logger
	db       *storage.PostgresStore
	client   *kgo.Client
	workChan chan Work
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
	w.workChan = make(chan Work, 100)

	for range 2 {
		w.workerWg.Add(1)
		go w.startShardWorker(ctx, w.workChan)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fetches := w.client.PollRecords(ctx, 50000)
			if fetches.IsClientClosed() {
				w.log.WarnContext(ctx, "Kafka client closed", slog.Any("ctx_err", ctx.Err()), slog.Any("client_err", w.client.Context().Err()))
				return nil
			}

			err := w.dispatchBatch(ctx, fetches)
			if err != nil {
				w.log.ErrorContext(ctx, "Failed to dispatch batch", slog.String("error", err.Error()))
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// TODO: manual commit offset to db
			// if err := w.client.CommitRecords(ctx, fetches.Records()...); err != nil {
			// 	w.log.ErrorContext(ctx, "Failed to commit records", slog.String("error", err.Error()))
			// }
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

func (w *Writer) startShardWorker(ctx context.Context, workChan chan Work) {
	defer w.workerWg.Done()
	for {
		select {
		case work, ok := <-workChan:
			if !ok {
				return
			}

			err := w.db.Transactions().WriteBatch(ctx, work.Records)
			work.Done <- err
		case <-time.After(20 * time.Second):
			w.log.DebugContext(ctx, "Worker idle")
		}
	}
}

func (w *Writer) dispatchBatch(ctx context.Context, fetches kgo.Fetches) error {
	records := fetches.Records()
	if len(records) == 0 {
		return nil
	}

	doneChan := make(chan error)

	w.workChan <- Work{
		Records: records,
		Done:    doneChan,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-doneChan:
		if err != nil {
			return err
		}
	}

	return nil
}
