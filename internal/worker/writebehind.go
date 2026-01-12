package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteBehindWorker struct {
	log          *slog.Logger
	pool         *pgxpool.Pool
	minPartition int
	maxPartition int
	idx          int
	cancel       context.CancelFunc
}

func NewWriteBehindWorker(log *slog.Logger, pool *pgxpool.Pool, minPartition int, maxPartition int) *WriteBehindWorker {
	return &WriteBehindWorker{
		log:          log,
		pool:         pool,
		minPartition: minPartition,
		maxPartition: maxPartition,
		idx:          minPartition,
	}
}
func (w *WriteBehindWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *WriteBehindWorker) run(ctx context.Context) error {
	var writeBehindCtx context.Context
	writeBehindCtx, w.cancel = context.WithCancel(ctx)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	w.log.Debug("Write behind worker started", slog.Int("min_partition", w.minPartition), slog.Int("max_partition", w.maxPartition))

	for {
		select {
		case <-writeBehindCtx.Done():
			w.log.Info("Write behind worker stopping")
			return nil
		case <-ticker.C:
			if err := w.writeBehind(w.idx); err != nil {
				w.log.Error("Write behind error", slog.Int("partition", w.idx), slog.Any("error", err))
			} else {
				w.log.Info("Write behind completed", slog.Int("partition", w.idx))
			}
			w.idx++
			if w.idx > w.maxPartition {
				w.idx = w.minPartition
			}
		}
	}
}

func (w *WriteBehindWorker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	return nil
}

func (w *WriteBehindWorker) writeBehind(i int) error {
	w.log.Debug("Writing behind for partition", slog.Int("partition", i))
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var exists bool
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM transactions_%d LIMIT 1)`, i)
	err := w.pool.QueryRow(timeoutCtx, query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existence of transactions for partition %d: %v", i, err)
	}

	if !exists {
		w.log.Debug("No transactions to write behind for partition", slog.Int("partition", i))
		return nil
	}

	tx, err := w.pool.Begin(timeoutCtx)
	defer tx.Rollback(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction for partition %d: %v", i, err)
	}

	update := fmt.Sprintf(`
		WITH aggregated_batch AS (
				SELECT 
						account_id, 
						SUM(amount) as net_change
				FROM transactions_%d
				GROUP BY account_id
		)
		UPDATE accounts
		SET balance = accounts.balance + aggregated_batch.net_change
		FROM aggregated_batch
		WHERE accounts.id = aggregated_batch.account_id;
	`, i)

	_, err = tx.Exec(timeoutCtx, update)
	if err != nil {
		return fmt.Errorf("failed to update accounts for partition %d: %v", i, err)
	}

	// Need a better archive strategy
	// archive := fmt.Sprintf(`
	// 	INSERT INTO transactions_history (id, account_id, amount, created_at)
	// 	SELECT id, account_id, amount, created_at FROM transactions_%d
	// 	ON CONFLICT (id) DO NOTHING;
	// `, i)
	// _, err = tx.Exec(timeoutCtx, archive)
	// if err != nil {
	// 	return fmt.Errorf("failed to archive transactions for partition %d: %v", i, err)
	// }

	clear := fmt.Sprintf(`TRUNCATE transactions_%d`, i)
	_, err = tx.Exec(timeoutCtx, clear)
	if err != nil {
		return fmt.Errorf("failed to clear transactions for partition %d: %v", i, err)
	}

	err = tx.Commit(timeoutCtx)
	if err != nil {
		return fmt.Errorf("failed to write behind for partition %d: %v", i, err)
	}
	return nil
}
