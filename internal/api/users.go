package api

import (
	"fmt"
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(p, "/")
		switch r.Method {
		case http.MethodGet:
			if len(parts) != 2 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Extract user ID from URL
			r.SetPathValue("userId", parts[1])
			s.handleGetUser()(w, r)
			return
		case http.MethodPost:
			s.handleCreateUser()(w, r)
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	// Id is the unique identifier of the user
	// @example 550e8400-e29b-41d4-a716-446655440000
	Id uuid.UUID `json:"id"`
	// CreatedAt is the timestamp when the user was created
	// @example 2025-12-25T11:11:00Z
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u *model.User) *UserResponse {
	return &UserResponse{
		Id:        u.Id,
		CreatedAt: time.UnixMilli(u.CreatedAt),
	}
}

// @Summary      Get user by ID
// @Description  Retrieves a user by their ID
// @Produce      plain
// @Param        id path int true "User ID"
// @Success      200 {object} UserResponse "User object"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /users/{id} [get]
func (s *Server) handleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := uuid.Parse(r.PathValue("userId"))
		if err != nil {
			s.respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		user, err := s.svc.Users.GetUser(r.Context(), userId)
		if err != nil {
			s.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		s.respondWithJSON(w, http.StatusOK, toUserResponse(user))
	}
}

// @Summary Create a new user
// @Description Creates a new user in the system
// @Produce plain
// @Success 201 {string} string "User created with ID"
// @Failure 500 {string} string "Failed to create user"
// @Router /users [post]
func (s *Server) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.svc.Users.CreateUser(r.Context())
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created with ID: %d", user.Id)
	}
}
