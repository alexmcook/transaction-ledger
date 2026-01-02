package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PartitionManager struct {
	log             *slog.Logger
	pool            *pgxpool.Pool
	activePartition int32
}

func NewPartitionManager(log *slog.Logger) *PartitionManager {
	return &PartitionManager{
		log: log,
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

func (pm *PartitionManager) StartRotationWorker(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := interval / 2

	go func() {
		for {
			select {
			case <-ticker.C:
				opCtx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				pm.rotateAndProcess(opCtx)
			case <-ctx.Done():
				pm.log.Info("Partition rotation worker stopping")
				return
			}
		}
	}()
}

func (pm *PartitionManager) rotateAndProcess(ctx context.Context) {
	partitionKey := pm.GetActivePartition()
	pm.switchPartition()
	pm.log.Info("Switched active partition", slog.Any("partition_key", partitionKey^1))

	// Wait for in flight transactions to complete
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		pm.log.Info("Partition processing cancelled before starting", slog.Any("partition_key", partitionKey))
		return
	}

	detachQuery := fmt.Sprintf(`ALTER TABLE transactions DETACH PARTITION transactions_p%d`, partitionKey)
	_, err := pm.pool.Exec(ctx, detachQuery)
	if err != nil {
		pm.log.Error("Failed to detach partition", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	tx, err := pm.pool.Begin(ctx)
	if err != nil {
		pm.log.Error("Failed to begin transaction for partition processing", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}
	defer tx.Rollback(ctx)

	sumQuery := fmt.Sprintf(`
		UPDATE accounts
		SET balance = balance + sub.total_amount
		FROM (
			SELECT account_id, SUM(amount) AS total_amount
			FROM transactions_p%d
			GROUP BY account_id
		) AS sub
		WHERE accounts.id = sub.account_id
		`,
		partitionKey,
	)

	_, err = tx.Exec(ctx, sumQuery)
	if err != nil {
		pm.log.Error("Failed to sum transactions and update account balances", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	truncateQuery := fmt.Sprintf(`TRUNCATE TABLE transactions_p%d`, partitionKey)
	_, err = tx.Exec(ctx, truncateQuery)
	if err != nil {
		pm.log.Error("Failed to truncate partition", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		pm.log.Error("Failed to commit partition processing transaction", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	attachQuery := fmt.Sprintf(`ALTER TABLE transactions ATTACH PARTITION transactions_p%d FOR VALUES IN (%d)`, partitionKey, partitionKey)
	_, err = pm.pool.Exec(ctx, attachQuery)
	if err != nil {
		pm.log.Error("Failed to attach partition", slog.Any("partition_key", partitionKey), slog.String("error", err.Error()))
		return
	}

	pm.log.Info("Successfully processed partition", slog.Any("partition_key", partitionKey))
}
