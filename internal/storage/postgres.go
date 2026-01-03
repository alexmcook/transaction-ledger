package storage

import (
	"log/slog"

	"github.com/alexmcook/transaction-ledger/internal/api"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	log              *slog.Logger
	pool             *pgxpool.Pool
	accountStore     *AccountStore
	transactionStore *TransactionStore
}

func NewPostgresStore(log *slog.Logger, pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		log:              log,
		pool:             pool,
		accountStore:     &AccountStore{pool: pool},
		transactionStore: &TransactionStore{pool: pool},
	}
}

func (ps *PostgresStore) Close() {
	ps.pool.Close()
}

func (ps *PostgresStore) Accounts() api.AccountStore {
	return ps.accountStore
}

func (ps *PostgresStore) Transactions() api.TransactionStore {
	return ps.transactionStore
}
