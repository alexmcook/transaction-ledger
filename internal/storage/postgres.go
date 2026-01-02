package storage

import (
	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartitionProvider interface {
	GetActivePartition() int16
}

type PostgresStore struct {
	pool             *pgxpool.Pool
	userStore        *UserStore
	accountStore     *AccountStore
	transactionStore *TransactionStore
}

func NewPostgresStore(pool *pgxpool.Pool, pp PartitionProvider) *PostgresStore {
	return &PostgresStore{
		pool:             pool,
		userStore:        &UserStore{pool: pool},
		accountStore:     &AccountStore{pool: pool},
		transactionStore: &TransactionStore{pool: pool, partitionProvder: pp},
	}
}

func (ps *PostgresStore) Close() {
	ps.pool.Close()
}

func (ps *PostgresStore) Users() api.UserStore {
	return ps.userStore
}

func (ps *PostgresStore) Accounts() api.AccountStore {
	return ps.accountStore
}

func (ps *PostgresStore) Transactions() api.TransactionStore {
	return ps.transactionStore
}
