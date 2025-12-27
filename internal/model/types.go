package model

import (
	"fmt"
)

type User struct {
	Id int64
	CreatedAt int64 // Milliseconds since epoch
}

type Account struct {
	Id int64
	UserId int64
	Balance int64
	CreatedAt int64 // Milliseconds since epoch
}

type Transaction struct {
	Id int64
	AccountId int64
	Amount int64
	CreatedAt int64 // Milliseconds since epoch
}

func (u *User) GetDetails() string {
	return fmt.Sprintf("User ID: %d\nCreated At: %d", u.Id, u.CreatedAt)
}

func (a *Account) GetDetails() string {
	return fmt.Sprintf("Account ID: %d\nUser ID: %d\nBalance: %d\nCreated At: %d", a.Id, a.UserId, a.Balance, a.CreatedAt)
}

func (t *Transaction) GetDetails() string {
	return fmt.Sprintf("Transaction ID: %d\nAccount ID: %d\nAmount: %d\nCreated At: %d", t.Id, t.AccountId, t.Amount, t.CreatedAt)
}
