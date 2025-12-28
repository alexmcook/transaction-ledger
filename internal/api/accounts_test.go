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

type MockAccountStore struct{}

func (m *MockAccountStore) GetAccount(ctx context.Context, id int64) (*model.Account, error) {
	return &model.Account{Id: id}, nil
}

func (m *MockAccountStore) CreateAccount(ctx context.Context, userId int64, balance int64) (*model.Account, error) {
	return &model.Account{Id: 1}, nil
}

func TestHandleCreateAccount(t *testing.T) {
	payload := []byte(`{"userId": 5, "balance": 100}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Accounts: &MockAccountStore{},
	}

	handler := handleCreateAccount(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", resp.StatusCode)
	}
}

func TestHandleGetAccount(t *testing.T) {
	// GET request for account with ID 1
	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	req.SetPathValue("accountId", "1") 
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Accounts: &MockAccountStore{},
	}

	handler := handleGetAccount(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleAccounts(t *testing.T) {
	var tests = []struct {
		name				 string
		method       string
		url          string
		body				 []byte
		expectedCode int
	}{
		{"GET", http.MethodGet, "/accounts", nil, http.StatusNoContent},
		{"GET", http.MethodGet, "/accounts/1", nil, http.StatusOK},
		{"POST", http.MethodPost, "/accounts", []byte(`{"userId": 5, "balance": 100}`), http.StatusCreated},
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
				Accounts: &MockAccountStore{},
			}

			handler := handleAccounts(svc)
			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("for %s %s: expected status %d, got %d", tt.method, tt.url, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
