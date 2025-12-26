package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type AccountsRepo struct {
	pool *pgxpool.Pool
}

func NewAccountsRepo(pool *pgxpool.Pool) *AccountsRepo {
	return &AccountsRepo{pool: pool}
}

func (r *AccountsRepo) GetUserAccounts(ctx context.Context, userId int64) ([]*model.Account, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account

	var account model.Account
	for rows.Next() {
		if err := rows.Scan(&account.Id, &account.UserId, &account.Balance, &account.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}

func (r *AccountsRepo) CreateAccount(ctx context.Context, userId int64, initialBalance int64) (*model.Account, error) {
	var account model.Account
	err := r.pool.QueryRow(ctx, "INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id, user_id, balance, created_at", userId, initialBalance).Scan(&account.Id, &account.UserId, &account.Balance, &account.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &account, nil
}
