package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartitionManager struct {
	log             *slog.Logger
	pool            *pgxpool.Pool
	activePartition int32
}

func NewPartitionManager(log *slog.Logger, pool *pgxpool.Pool) *PartitionManager {
	return &PartitionManager{
		log:  log,
		pool: pool,
	}
}

func (pm *PartitionManager) GetActivePartition() int16 {
	return int16(atomic.LoadInt32(&pm.activePartition))
}

func (pm *PartitionManager) switchPartition() {
	current := atomic.LoadInt32(&pm.activePartition)
	next := current ^ 1 // Toggle partition between 0 and 1
	if atomic.CompareAndSwapInt32(&pm.activePartition, current, next) {
		return
	}
}

func (pm *PartitionManager) StartRotationWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	timeout := interval / 2

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				opCtx, cancel := context.WithTimeout(context.Background(), timeout)
				pm.rotateAndProcess(opCtx)
				cancel()
			case <-ctx.Done():
				pm.log.InfoContext(ctx, "Partition rotation worker stopping")
				return
			}
		}
	}()
}

func (pm *PartitionManager) rotateAndProcess(ctx context.Context) {
	partitionKey := pm.GetActivePartition()
	pm.switchPartition()
	pm.log.InfoContext(ctx, "Switched active partition", slog.Any("partition_key", partitionKey^1))

	// Wait for in flight transactions to complete
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		pm.log.InfoContext(ctx, "Partition processing cancelled before starting", slog.Any("partition_key", partitionKey))
		return
	}

	err := pgx.BeginFunc(ctx, pm.pool, func(tx pgx.Tx) error {
		// Detach
		// Locks parent table for duration of transaction, can optimize by detaching before and handling orphaned partitions on startup
		detachQuery := fmt.Sprintf("ALTER TABLE transactions DETACH PARTITION transactions_p%d", partitionKey)
		_, err := tx.Exec(ctx, detachQuery)
		if err != nil {
			return err
		}

		// Sum and update
		updateQuery := fmt.Sprintf(`
			UPDATE accounts a
			SET balance = a.balance + sub.total_amount
			FROM (
				SELECT account_id, SUM(amount) AS total_amount
				FROM transactions_p%d
				GROUP BY account_id
			) AS sub
			WHERE a.id = sub.account_id
		`, partitionKey)
		_, err = tx.Exec(ctx, updateQuery)
		if err != nil {
			return err
		}

		// Truncate
		truncateQuery := fmt.Sprintf("TRUNCATE TABLE transactions_p%d", partitionKey)
		_, err = tx.Exec(ctx, truncateQuery)
		if err != nil {
			return err
		}

		// Reattach
		reattachQuery := fmt.Sprintf("ALTER TABLE transactions ATTACH PARTITION transactions_p%d FOR VALUES IN (%d)", partitionKey, partitionKey)
		_, err = tx.Exec(ctx, reattachQuery)
		return err
	})

	if err != nil {
		pm.log.ErrorContext(ctx, "Error processing partition", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	pm.log.InfoContext(ctx, "Successfully processed partition", slog.Any("partition_key", partitionKey))
}
