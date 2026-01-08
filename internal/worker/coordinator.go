package worker

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/alexmcook/transaction-ledger/internal/storage"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Coordinator struct {
	log     *slog.Logger
	client  *kgo.Client
	workers []*MultiWriter
}

func NewCoordinator(ctx context.Context, minPart int, maxPart int, log *slog.Logger, db *storage.PostgresStore, client *kgo.Client) *Coordinator {
	numWorkers := maxPart - minPart + 1

	c := &Coordinator{
		log:     log,
		client:  client,
		workers: make([]*MultiWriter, numWorkers),
	}

	for i := range c.workers {
		c.workers[i] = NewMultiWriter(minPart+i, log, db)
	}

	return c
}

func (c *Coordinator) Run(ctx context.Context) error {
	workerCtx := context.WithoutCancel(ctx)
	for _, w := range c.workers {
		w.Start(workerCtx)
	}

	numWorkers := len(c.workers)
	activeSlabs := make([]*RecordBatch, numWorkers)

	for i := range activeSlabs {
		activeSlabs[i] = recordsPool.Get().(*RecordBatch)
		activeSlabs[i].Reset()
	}

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			c.log.WarnContext(ctx, "Kafka client closed", slog.Any("ctx_err", ctx.Err()), slog.Any("client_err", c.client.Context().Err()))
			// TODO: retry reconnect
			return nil
		}

		iter := fetches.RecordIter()

		for !iter.Done() {
			rec := iter.Next()
			workerID := int(rec.Partition) % numWorkers
			batch := activeSlabs[workerID]

			dest := &batch.Slab[batch.Count]
			*dest = *rec
			copy(batch.ByteSlab[batch.offset:], rec.Value)
			dest.Value = batch.ByteSlab[batch.offset : batch.offset+len(rec.Value)]
			batch.offset += len(rec.Value)
			batch.Count++

			if batch.Count >= 50000 {
				c.dispatch(workerID, batch)
				newSlab := recordsPool.Get().(*RecordBatch)
				activeSlabs[workerID] = newSlab
			}
		}

		for i, batch := range activeSlabs {
			if batch.Count > 0 {
				c.dispatch(i, batch)
				activeSlabs[i] = recordsPool.Get().(*RecordBatch)
				activeSlabs[i].Reset()
			}
		}
	}
}

func (c *Coordinator) Stop(ctx context.Context) error {
	for _, w := range c.workers {
		if err := w.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) dispatch(workerID int, batch *RecordBatch) {
	lastOffset := batch.Slab[batch.Count-1].Offset
	workerIDStr := strconv.Itoa(workerID)
	kafkaHighWatermark.WithLabelValues(workerIDStr).Set(float64(lastOffset))

	c.workers[workerID].WorkChan <- batch
}
