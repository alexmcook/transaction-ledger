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
	GetUserAccounts(ctx context.Context, userId int64) ([]*model.Account, error)
	CreateAccount(ctx context.Context, userId int64, initialBalance int64) (*model.Account, error)
}

type Service struct {
	Users    UserStore
	Accounts AccountStore
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
