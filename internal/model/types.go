package model

import (
	"fmt"
	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID
	CreatedAt int64 // Milliseconds since epoch
}

type Account struct {
	Id uuid.UUID
	UserId uuid.UUID
	Balance int64
	CreatedAt int64 // Milliseconds since epoch
}

type Transaction struct {
	Id uuid.UUID
	AccountId uuid.UUID
	Amount int64
	CreatedAt int64 // Milliseconds since epoch
}

func (u *User) GetDetails() string {
	return fmt.Sprintf("User ID: %s\nCreated At: %d", u.Id, u.CreatedAt)
}

func (a *Account) GetDetails() string {
	return fmt.Sprintf("Account ID: %s\nUser ID: %s\nBalance: %d\nCreated At: %d", a.Id, a.UserId, a.Balance, a.CreatedAt)
}

func (t *Transaction) GetDetails() string {
	return fmt.Sprintf("Transaction ID: %s\nAccount ID: %s\nAmount: %d\nCreated At: %d", t.Id, t.AccountId, t.Amount, t.CreatedAt)
}
