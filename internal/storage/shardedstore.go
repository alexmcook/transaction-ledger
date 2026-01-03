package storage

import (
	"context"
	"encoding/binary"
	"log/slog"

	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShardedStore struct {
	log    *slog.Logger
	shards []*PostgresStore
}

func NewShardedStore(log *slog.Logger, pools []*pgxpool.Pool) *ShardedStore {
	shards := make([]*PostgresStore, len(pools))
	for i, pool := range pools {
		shards[i] = NewPostgresStore(log, pool)
	}
	return &ShardedStore{
		log:    log,
		shards: shards,
	}
}

func (s *ShardedStore) getShard(uid uuid.UUID) *PostgresStore {
	// Entropy from last 8 bytes
	val := binary.BigEndian.Uint64(uid[8:16])
	shardKey := val % uint64(len(s.shards))
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
