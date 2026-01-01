package api

import (
	"github.com/gofiber/fiber/v3"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	s := NewServer(nil, slog.Default())

	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	expectedBody := "OK"
	body, _ := io.ReadAll(resp.Body)
	if string(body) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}
