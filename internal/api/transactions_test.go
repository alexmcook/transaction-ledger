package api

import (
	"fmt"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"github.com/google/uuid"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type MockTransactionStore struct{}

func (m *MockTransactionStore) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return &model.Transaction{Id: id}, nil
}

func (m *MockTransactionStore) CreateTransaction(ctx context.Context, accountId uuid.UUID, amount int64) (*model.Transaction, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &model.Transaction{Id: uuid}, nil
}

func TestHandleCreateTransaction(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	payload := fmt.Appendf(nil, `{"accountId": "%s", "amount": 50}`, uuid)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Transactions: &MockTransactionStore{},
	}

	handler := handleCreateTransaction(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", resp.StatusCode)
	}
}

func TestHandleGetTransaction(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/transactions/" + uuid.String(), nil)
	req.SetPathValue("transactionId", uuid.String()) 
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Transactions: &MockTransactionStore{},
	}

	handler := handleGetTransaction(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleTransactions(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	var tests = []struct {
		name				 string
		method       string
		url          string
		body				 []byte
		expectedCode int
	}{
		{"GET", http.MethodGet, "/transactions", nil, http.StatusNoContent},
		{"GET", http.MethodGet, "/transactions/" + uuid.String(), nil, http.StatusOK},
		{"POST", http.MethodPost, "/transactions", fmt.Appendf(nil, `{"accountId": "%s", "amount": 50}`, uuid), http.StatusCreated},
	}

	for _, tt := range tests {
		t. Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body == nil {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			} else {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			// Mock service
			svc := &service.Service{
				Transactions: &MockTransactionStore{},
			}

			handler := handleTransactions(svc)
			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("for %s %s: expected status %d, got %d", tt.method, tt.url, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
