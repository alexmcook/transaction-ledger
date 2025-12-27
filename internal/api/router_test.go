package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"testing"
	"context"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}

	expectedBody := "OK"
	body := w.Body.String()
	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

type MockUserStore struct{}

func (m *MockUserStore) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return &model.User{Id: id}, nil
}

func (m *MockUserStore) CreateUser(ctx context.Context) (*model.User, error) {
	return &model.User{Id: 1}, nil
}

func TestHandleCreateUser(t *testing.T) {
	data := struct {
		Name string `json:"name"`
	}{
		Name: "alex",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	reqBody := bytes.NewBuffer(jsonData)

	// POST request with value "alex"
	req := httptest.NewRequest(http.MethodPost, "/users", reqBody)
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
	// GET request for user with ID 1
	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.SetPathValue("userId", "1") 
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

	expectedBody := "User ID: 1"
	body := w.Body.String()
	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestHandleUsers(t *testing.T) {
	var tests = []struct {
		name				 string
		method       string
		url          string
		expectedCode int
	}{
		{"GET", http.MethodGet, "/users", http.StatusNoContent},
		{"GET", http.MethodGet, "/users/1", http.StatusOK},
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
