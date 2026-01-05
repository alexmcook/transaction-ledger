package worker

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Writer struct {
	log        *slog.Logger
	shards     *storage.ShardedStore
	client     *kgo.Client
	numShards  int
	shardChans []chan ShardWork
	workerWg   sync.WaitGroup
}

func NewWriter(log *slog.Logger, shards *storage.ShardedStore, client *kgo.Client, numShards int) *Writer {
	return &Writer{
		log:       log,
		shards:    shards,
		client:    client,
		numShards: numShards,
	}
}

func (w *Writer) getShard(key []byte) int {
	// Entropy from last 8 bytes
	val := binary.BigEndian.Uint64(key[8:16])
	return int(val % uint64(w.numShards))
}

func (w *Writer) Start(ctx context.Context) error {
	w.shardChans = make([]chan ShardWork, w.numShards)

	for i := range w.numShards {
		w.shardChans[i] = make(chan ShardWork, 2)
		for range 2 {
			w.workerWg.Add(1)
			go w.startShardWorker(ctx, i, w.shardChans[i])
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fetches := w.client.PollRecords(ctx, 50000)
			if fetches.IsClientClosed() {
				return nil
			}

			err := w.dispatchBatch(ctx, fetches)
			if err != nil {
				w.log.ErrorContext(ctx, "Failed to dispatch batch", slog.String("error", err.Error()))
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := w.client.CommitRecords(ctx, fetches.Records()...); err != nil {
				w.log.ErrorContext(ctx, "Failed to commit records", slog.String("error", err.Error()))
			}
		}
	}
}

func (w *Writer) Stop(ctx context.Context) error {
	for _, ch := range w.shardChans {
		close(ch)
	}

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

func (w *Writer) startShardWorker(ctx context.Context, shardID int, workChan chan ShardWork) {
	defer w.workerWg.Done()
	for {
		select {
		case work, ok := <-workChan:
			if !ok {
				return
			}

			err := w.shards.WriteBatch(ctx, shardID, work.Records)
			work.Done <- err
		case <-time.After(10 * time.Second):
			w.log.DebugContext(ctx, "Shard worker idle", slog.Int("shardID", shardID))
		}
	}
}

func (w *Writer) dispatchBatch(ctx context.Context, fetches kgo.Fetches) error {
	records := fetches.Records()
	if len(records) == 0 {
		return nil
	}

	shards := make([][]*kgo.Record, w.numShards)
	for _, record := range records {
		shardID := w.getShard(record.Key)
		shards[shardID] = append(shards[shardID], record)
	}

	shardsToAck := 0
	doneChan := make(chan error, w.numShards)

	for shardID, shardRecords := range shards {
		if len(shardRecords) == 0 {
			continue
		}
		shardsToAck++

		w.shardChans[shardID] <- ShardWork{
			Records: shardRecords,
			Done:    doneChan,
		}
	}

	for range shardsToAck {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-doneChan:
			if err != nil {
				return err
			}
		}
	}

	return nil
}
