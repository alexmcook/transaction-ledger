package model

import (
	"fmt"
	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID
	CreatedAt int64 // Milliseconds since epoch
}

type Account struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Balance   int64
	CreatedAt int64 // Milliseconds since epoch
}

type Transaction struct {
	Id        uuid.UUID
	AccountId uuid.UUID
	Amount    int64
	Type      int   // Credit or Debit
	CreatedAt int64 // Milliseconds since epoch
}

// TransactionPayload represents the transaction data received in API requests
type TransactionPayload struct {
	// AccountId is the unique identifier of the account associated with the transaction
	//	@example	880e8400-e29b-41d4-a716-446655440000
	AccountId uuid.UUID `json:"accountId" binding:"required"`
	// Type is the type of the transaction (e.g., credit or debit) as an integer
	//	@example	0
	Type int `json:"type" binding:"required"`
	// Amount is the amount of the transaction
	//	@example	500
	Amount int64 `json:"amount" binding:"required"`
}

const (
	Credit = iota
	Debit
)

func (u *User) GetDetails() string {
	return fmt.Sprintf("User ID: %s\nCreated At: %d", u.Id, u.CreatedAt)
}

func (a *Account) GetDetails() string {
	return fmt.Sprintf("Account ID: %s\nUser ID: %s\nBalance: %d\nCreated At: %d", a.Id, a.UserId, a.Balance, a.CreatedAt)
}

func (t *Transaction) GetDetails() string {
	var typeStr string
	switch t.Type {
	case Credit:
		typeStr = "Credit"
	case Debit:
		typeStr = "Debit"
	}
	return fmt.Sprintf("Transaction ID: %s\nAccount ID: %s\nAmount: %d\nType: %s\nCreated At: %d", t.Id, t.AccountId, t.Amount, typeStr, t.CreatedAt)
}
