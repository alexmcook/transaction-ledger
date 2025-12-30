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

type MockTransactionStore struct{}

func (m *MockTransactionStore) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return &model.Transaction{Id: id}, nil
}

func (m *MockTransactionStore) CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64) (*model.Transaction, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &model.Transaction{Id: uuid}, nil
}

func TestHandleGetTransaction(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/transactions/"+uuid.String(), nil)
	req.SetPathValue("transactionId", uuid.String())
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Transactions: &MockTransactionStore{},
	}

	s := &Server{
		logger: logger.Init(false),
		svc:    svc,
	}

	handler := s.handleGetTransaction()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleCreateTransaction(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	var tests = []struct {
		name         string
		payload      []byte
		expectedCode int
	}{
		{"ValidTransaction", fmt.Appendf(nil, `{"accountId":"%s","type":1,"amount":1000}`, uuid.String()), http.StatusCreated},
		{"InvalidAccountId", []byte(`{"accountId":"invalid-uuid","type":1,"amount":1000}`), http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Mock service
			svc := &service.Service{
				Transactions: &MockTransactionStore{},
				Accounts:     &MockAccountStore{},
			}

			s := &Server{
				logger: logger.Init(false),
				svc:    svc,
			}

			handler := s.handleCreateTransaction()
			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
