package api

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestHandleGetUser(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.String(), nil)
	req.SetPathValue("userId", uuid.String())
	w := httptest.NewRecorder()

	// Mock service
	svc := &service.Service{
		Users: &MockUserStore{},
	}

	logger, err := logger.Init(false)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	s := &Server{
		logger: logger,
		svc:    svc,
	}

	handler := s.handleGetUser()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleCreateUser(t *testing.T) {
	var tests = []struct {
		name         string
		url          string
		expectedCode int
	}{
		{"Valid", "/users", http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			// Mock service
			svc := &service.Service{
				Users: &MockUserStore{},
			}

			logger, err := logger.Init(false)
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			s := &Server{
				logger: logger,
				svc:    svc,
			}

			handler := s.handleCreateUser()
			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
