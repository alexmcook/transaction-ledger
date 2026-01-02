package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Account struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Balance   int64     `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AccountID       uuid.UUID `json:"account_id" db:"account_id"`
	Amount          int64     `json:"amount" db:"amount"`
	TransactionType int8      `json:"transaction_type" db:"type"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
