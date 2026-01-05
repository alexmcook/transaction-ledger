package storage

import (
	"context"
	"encoding/binary"
	"log/slog"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ShardedStore struct {
	log       *slog.Logger
	shards    []*PostgresStore
	numShards int
}

func NewShardedStore(log *slog.Logger, pools []*pgxpool.Pool) *ShardedStore {
	shards := make([]*PostgresStore, len(pools))
	for i, pool := range pools {
		shards[i] = NewPostgresStore(log, pool)
	}
	return &ShardedStore{
		log:       log,
		shards:    shards,
		numShards: len(pools),
	}
}

func (s *ShardedStore) getShard(uid uuid.UUID) *PostgresStore {
	// Entropy from last 8 bytes
	val := binary.BigEndian.Uint64(uid[8:16])
	shardKey := val % uint64(s.numShards)
	return s.shards[shardKey]
}

func (s *ShardedStore) GetAccount(ctx context.Context, uid uuid.UUID) (*model.Account, error) {
	shard := s.getShard(uid)
	return shard.Accounts().GetAccount(ctx, uid)
}

func (s *ShardedStore) GetTransaction(ctx context.Context, uid uuid.UUID) (*model.Transaction, error) {
	shard := s.getShard(uid)
	return shard.Transactions().GetTransaction(ctx, uid)
}

func (s *ShardedStore) WriteBatch(ctx context.Context, shardId int, batch []*kgo.Record) error {
	s.log.DebugContext(ctx, "Writing batch to shard", slog.Int("shardId", shardId), slog.Int("batchSize", len(batch)))
	return s.shards[shardId].transactionStore.WriteBatch(ctx, shardId, batch)
}
