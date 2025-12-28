package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"github.com/google/uuid"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

type MockUserStore struct{}

func (m *MockUserStore) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return &model.User{Id: id}, nil
}

func (m *MockUserStore) CreateUser(ctx context.Context) (*model.User, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &model.User{Id: uuid}, nil
}

func TestHandleCreateUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Users: &MockUserStore{},
	}

	handler := handleCreateUser(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", resp.StatusCode)
	}
}

func TestHandleGetUser(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/" + uuid.String(), nil)
	req.SetPathValue("userId", uuid.String()) 
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Users: &MockUserStore{},
	}

	handler := handleGetUser(svc)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleUsers(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	var tests = []struct {
		name				 string
		method       string
		url          string
		expectedCode int
	}{
		{"GET", http.MethodGet, "/users", http.StatusNoContent},
		{"GET", http.MethodGet, "/users/" + uuid.String(), http.StatusOK},
		{"POST", http.MethodPost, "/users", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			// Mock service
			svc := &service.Service{
				Users: &MockUserStore{},
			}

			handler := handleUsers(svc)
			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
