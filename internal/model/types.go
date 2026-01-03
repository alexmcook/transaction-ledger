package model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Balance   int64     `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AccountID uuid.UUID `json:"account_id" db:"account_id"`
	Amount    int64     `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
