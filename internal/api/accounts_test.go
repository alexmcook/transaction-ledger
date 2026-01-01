package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"log/slog"
	"net/http/httptest"
	"testing"
)

type MockAccountStore struct{}

func (m *MockAccountStore) GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	return &model.Account{Id: id}, nil
}

func (m *MockAccountStore) CreateAccount(ctx context.Context, userId uuid.UUID, balance int64) (*model.Account, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &model.Account{Id: uuid}, nil
}

func (m *MockAccountStore) UpdateAccountBalance(ctx context.Context, accountId uuid.UUID, amount int64) error {
	return nil
}

func TestHandleGetAccount(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	// Mock service
	svc := &service.Service{
		Accounts: &MockAccountStore{},
	}

	s := NewServer(svc, slog.Default())

	target := "/accounts/" + uuid.String()
	req := httptest.NewRequest(fiber.MethodGet, target, nil)

	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	expectedType := "application/json; charset=utf-8"
	if contentType != expectedType {
		t.Errorf("expected Content-Type %q, got %q", expectedType, contentType)
	}
}

func TestHandleCreateAccount(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	var tests = []struct {
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:         "ValidAccount",
			payload:      fmt.Sprintf(`{"userId": "%s", "balance": %d}`, uuid.String(), 100),
			expectedCode: fiber.StatusCreated,
		},
		{
			name:         "InvalidUserId",
			payload:      `{"userId": "invalid-uuid", "balance": 100}`,
			expectedCode: fiber.StatusBadRequest,
		},
	}

	// Mock service
	svc := &service.Service{
		Accounts: &MockAccountStore{},
	}

	s := NewServer(svc, slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/accounts", bytes.NewReader([]byte(tt.payload)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to perform request: %v", err)
			}
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			expectedType := "application/json; charset=utf-8"
			if contentType != expectedType {
				t.Errorf("test %q: expected Content-Type %q, got %q", tt.name, expectedType, contentType)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("test %q: expected status %d, got %d", tt.name, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
