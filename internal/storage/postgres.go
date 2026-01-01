package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/api"
)

type PostgresStore struct {
	pool  *pgxpool.Pool
	userStore *UserStore
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		pool:  pool,
		userStore: &UserStore{pool: pool},
	}
}

func (ps *PostgresStore) Close() {
	ps.pool.Close()
}

func (ps *PostgresStore) Users() api.UserStore {
	return ps.userStore
}
