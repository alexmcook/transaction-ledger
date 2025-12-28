package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/alexmcook/transaction-ledger/internal/model"
)

type MockTransactionStore struct{}

func (m *MockTransactionStore) GetTransaction(ctx context.Context, id int64) (*model.Transaction, error) {
	return &model.Transaction{Id: id}, nil
}

func (m *MockTransactionStore) CreateTransaction(ctx context.Context, accountId int64, amount int64) (*model.Transaction, error) {
	return &model.Transaction{Id: 1}, nil
}

func TestHandleCreateTransaction(t *testing.T) {
	payload := []byte(`{"accountId": 5, "amount": 50}`)
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
	// GET request for transaction with ID 1
	req := httptest.NewRequest(http.MethodGet, "/transactions/1", nil)
	req.SetPathValue("transactionId", "1") 
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
	var tests = []struct {
		name				 string
		method       string
		url          string
		body				 []byte
		expectedCode int
	}{
		{"GET", http.MethodGet, "/transactions", nil, http.StatusNoContent},
		{"GET", http.MethodGet, "/transactions/1", nil, http.StatusOK},
		{"POST", http.MethodPost, "/transactions", []byte(`{"accountId": 5, "amount": 50}`), http.StatusCreated},
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
