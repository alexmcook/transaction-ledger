package api

import (
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/google/uuid"
	"net/http"
	"time"
)

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
			s.respondWithError(r.Context(), w, http.StatusBadRequest, "Invalid user ID format", err)
			return
		}

		user, err := s.svc.Users.GetUser(r.Context(), userId)
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusNotFound, "User not found", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusOK, toUserResponse(user))
	}
}

// @Summary      Create a new user
// @Description  Creates a new user in the system
// @Produce      json
// @Success      201 {object} UserResponse "User object"
// @Failure      500 {object} ErrorResponse "Failed to create user"
// @Router       /users [post]
func (s *Server) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.svc.Users.CreateUser(r.Context())
		if err != nil {
			s.respondWithError(r.Context(), w, http.StatusInternalServerError, "Failed to create user", err)
			return
		}

		s.respondWithJSON(r.Context(), w, http.StatusCreated, toUserResponse(user))
	}
}
