package service

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type UserStore interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
	CreateUser(ctx context.Context) (*model.User, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, id int64) (*model.Account, error)
	CreateAccount(ctx context.Context, userId int64, initialBalance int64) (*model.Account, error)
}

type TransactionStore interface {
	GetTransaction(ctx context.Context, id int64) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, accountId int64, amount int64) (*model.Transaction, error)
}

type Service struct {
	Users       	UserStore
	Accounts    	AccountStore
	Transactions	TransactionStore
}

type Deps struct {
	Users    UserStore
	Accounts AccountStore
}

func New(d Deps) *Service {
	return &Service{
		Users:    d.Users,
		Accounts: d.Accounts,
	}
}
