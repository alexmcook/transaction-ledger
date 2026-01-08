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

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			c.log.WarnContext(ctx, "Kafka client closed", slog.Any("ctx_err", ctx.Err()), slog.Any("client_err", c.client.Context().Err()))
			// TODO: retry reconnect
			return nil
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, f := range errs {
				c.log.ErrorContext(ctx, "Fetch error", slog.Int("partition", int(f.Partition)), slog.String("topic", f.Topic), slog.Any("error", f.Err))
			}
		}

		if fetches.Empty() {
			continue
		}

		fetches.EachPartition(func(fp kgo.FetchTopicPartition) {
			workerID := int(fp.Partition) % numWorkers

			highwater := fp.HighWatermark
			partiotionStr := strconv.Itoa(int(fp.Partition))
			kafkaHighWatermark.WithLabelValues(partiotionStr).Set(float64(highwater))

			fpCopy := fp
			select {
			case c.workers[workerID].WorkChan <- &fpCopy:
			case <-ctx.Done():
				return
			}
		})
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
