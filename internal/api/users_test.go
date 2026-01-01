package api

import (
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"log/slog"
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

	// Mock service
	svc := &service.Service{
		Users: &MockUserStore{},
	}

	s := NewServer(svc, slog.Default())

	target := "/users/" + uuid.String()
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
	expectedContentType := "application/json; charset=utf-8"
	if contentType != expectedContentType {
		t.Errorf("expected Content-Type %q, got %q", expectedContentType, contentType)
	}
}

func TestHandleCreateUser(t *testing.T) {
	var tests = []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "Valid",
			expectedCode: fiber.StatusCreated,
		},
	}

	// Mock service
	svc := &service.Service{
		Users: &MockUserStore{},
	}

	s := NewServer(svc, slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/users", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("failed to perform request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			expectedContentType := "application/json; charset=utf-8"
			if contentType != expectedContentType {
				t.Errorf("expected Content-Type %q, got %q", expectedContentType, contentType)
			}
		})
	}
}
