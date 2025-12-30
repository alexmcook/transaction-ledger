package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/google/uuid"
	"net/http"
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

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+uuid.String(), nil)
	req.SetPathValue("accountId", uuid.String())
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Accounts: &MockAccountStore{},
	}

	logger, err := logger.Init(false)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	s := &Server{
		logger: logger,
		svc:    svc,
	}

	handler := s.handleGetAccount()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleCreateAccount(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	var tests = []struct {
		name         string
		url          string
		payload      []byte
		expectedCode int
	}{
		{"ValidAccount", "/accounts", fmt.Appendf(nil, `{"userId": "%s", "balance": %d}`, uuid.String(), 100), http.StatusCreated},
		{"InvalidUserId", "/accounts", []byte(`{"userId": "invalid-uuid", "balance": 100}`), http.StatusBadRequest},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(tt.payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Mock service
		svc := &service.Service{
			Accounts: &MockAccountStore{},
		}

		logger, err := logger.Init(false)
		if err != nil {
			t.Fatalf("Failed to initialize logger: %v", err)
		}

		s := &Server{
			logger: logger,
			svc:    svc,
		}

		handler := s.handleCreateAccount()
		handler(w, req)

		resp := w.Result()
		if resp.StatusCode != tt.expectedCode {
			t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
		}
	}
}
