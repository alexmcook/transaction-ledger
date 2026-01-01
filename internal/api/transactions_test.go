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

type MockTransactionStore struct{}

func (m *MockTransactionStore) GetTransaction(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return &model.Transaction{Id: id}, nil
}

func (m *MockTransactionStore) CreateTransaction(ctx context.Context, accountId uuid.UUID, transactionType int, amount int64, bucketId int32) (*model.Transaction, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &model.Transaction{Id: uuid}, nil
}

type MockBucketProvider struct{}

func (fw *MockBucketProvider) GetActiveBucket() int32 {
	return 0
}

func TestHandleGetTransaction(t *testing.T) {
	uuid, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	// Mock service
	svc := &service.Service{
		Transactions: &MockTransactionStore{},
	}

	s := NewServer(svc, slog.Default())

	target := "/transactions/" + uuid.String()
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

func TestHandleCreateTransaction(t *testing.T) {
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
			name:         "ValidTransaction",
			payload:      fmt.Sprintf(`{"accountId":"%s","type":1,"amount":1000}`, uuid.String()),
			expectedCode: fiber.StatusCreated,
		},
		{
			name:         "InvalidAccountId",
			payload:      `{"accountId":"invalid-uuid","type":1,"amount":1000}`,
			expectedCode: fiber.StatusBadRequest,
		},
	}

	// Mock service
	svc := &service.Service{
		Transactions:   &MockTransactionStore{},
		Accounts:       &MockAccountStore{},
		BucketProvider: &MockBucketProvider{},
		TxChan:         make(chan *model.Transaction, 100),
	}

	s := NewServer(svc, slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/transactions", bytes.NewReader([]byte(tt.payload)))
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
			expectedType := "application/json; charset=utf-8"
			if contentType != expectedType {
				t.Errorf("expected Content-Type %q, got %q", expectedType, contentType)
			}
		})
	}
}
