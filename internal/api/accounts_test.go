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

func TestHandleCreateAccount(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	payload := fmt.Appendf(nil, `{"userId": "%s", "balance": 100}`, uuid)
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
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/accounts/" + uuid.String(), nil)
	req.SetPathValue("accountId", uuid.String()) 
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
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}

	var tests = []struct {
		name				 string
		method       string
		url          string
		body				 []byte
		expectedCode int
	}{
		{"GET", http.MethodGet, "/accounts", nil, http.StatusNoContent},
		{"GET", http.MethodGet, "/accounts/" + uuid.String(), nil, http.StatusOK},
		{"POST", http.MethodPost, "/accounts", fmt.Appendf(nil, `{"userId": "%s", "balance": 100}`, uuid), http.StatusCreated},
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
