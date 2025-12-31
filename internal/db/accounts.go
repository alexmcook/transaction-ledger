package db

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AccountsRepo struct {
	pool *pgxpool.Pool
}

func NewAccountsRepo(pool *pgxpool.Pool) *AccountsRepo {
	return &AccountsRepo{pool: pool}
}

func (r *AccountsRepo) GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	var account model.Account

	query := `
		SELECT
			a.id, 
			a.user_id,
			a.balance + COALESCE(SUM(t.amount), 0) AS balance,
			a.created_at
		FROM accounts a
		LEFT JOIN transactions t ON t.account_id = a.id
		WHERE a.id = $1
		GROUP BY a.id, a.user_id, a.balance, a.created_at
	`

	err := r.pool.
		QueryRow(ctx, query, id).
		Scan(&account.Id, &account.UserId, &account.Balance, &account.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AccountsRepo) CreateAccount(ctx context.Context, userId uuid.UUID, initialBalance int64) (*model.Account, error) {
	var account model.Account
	accountId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	account.Id = accountId
	account.UserId = userId
	account.Balance = initialBalance
	account.CreatedAt = time.Now().UnixMilli()

	_, err = r.pool.Exec(ctx, "INSERT INTO accounts (id, user_id, balance, created_at) VALUES ($1, $2, $3, $4)", account.Id, account.UserId, account.Balance, account.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountsRepo) UpdateAccountBalance(ctx context.Context, accountId uuid.UUID, amount int64) error {
	_, err := r.pool.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, accountId)
	return err
}

func (r *AccountsRepo) UpdateBalances(ctx context.Context, balances map[uuid.UUID]int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for accountId, amount := range balances {
		_, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, accountId)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
