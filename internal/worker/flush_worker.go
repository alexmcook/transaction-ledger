package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Flushable interface {
	FlushBucket(ctx context.Context, bucketId int32) error
}

type FlushWorker struct {
	flushInterval time.Duration
	logger        *slog.Logger
	flushable     Flushable
	activeBucket  int32
	workerOnce    sync.Once
}

type FlushWorkerOpts struct {
	FlushInterval time.Duration
	Logger        *slog.Logger
	Flushable     Flushable
}

func NewFlushWorker(opts FlushWorkerOpts) *FlushWorker {
	return &FlushWorker{
		flushInterval: opts.FlushInterval,
		logger:        opts.Logger,
		flushable:     opts.Flushable,
		activeBucket:  0,
	}
}

func (w *FlushWorker) flushRoutine(ctx context.Context) {
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bucketId := w.GetActiveBucket()
			w.logger.Info("FlushWorker: flushing write-behind buffer", slog.Int("bucket", int(bucketId)))

			w.switchActiveBucket() // Switch partitions

			select {
			case <-time.After(100 * time.Millisecond):
				// Ensure any in-flight transactions complete
			case <-ctx.Done():
				return
			}

			err := w.flushable.FlushBucket(ctx, bucketId) // Flush the inactive bucket
			if err != nil {
				w.logger.Error("FlushWorker: error flushing write-behind buffer", slog.Int("bucket", int(bucketId)), slog.String("error", err.Error()))
			} else {
				w.logger.Info("FlushWorker: successfully flushed write-behind buffer", slog.Int("bucket", int(bucketId)))
			}
		case <-ctx.Done():
			w.logger.Info("FlushWorker: stopping flush worker")
			return
		}
	}
}

func (w *FlushWorker) Start(ctx context.Context) {
	w.workerOnce.Do(func() {
		w.logger.Info("FlushWorker: starting flush worker", slog.Duration("flush_interval", w.flushInterval))
		go w.flushRoutine(ctx)
	})
}

func (w *FlushWorker) GetActiveBucket() int32 {
	return atomic.LoadInt32(&w.activeBucket)
}

func (w *FlushWorker) switchActiveBucket() {
	// Atomically toggle between 0 and 1
	for {
		current := atomic.LoadInt32(&w.activeBucket)
		if atomic.CompareAndSwapInt32(&w.activeBucket, current, 1-current) {
			return
		}
	}
}
