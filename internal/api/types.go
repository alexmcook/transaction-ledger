package api

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"time"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type CreateUserRequest struct {
	ID uuid.UUID `json:"id"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreRegistry interface {
	Users() UserStore
}

type UserStore interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateUser(ctx context.Context, id uuid.UUID, createdAt time.Time) error
}
