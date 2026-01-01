package api

import (
	"github.com/alexmcook/transaction-ledger/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	// Id is the unique identifier of the user
	//	@example	550e8400-e29b-41d4-a716-446655440000
	Id uuid.UUID `json:"id"`
	// CreatedAt is the timestamp when the user was created
	//	@example	2025-12-25T11:11:00Z
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u *model.User) *UserResponse {
	return &UserResponse{
		Id:        u.Id,
		CreatedAt: time.UnixMilli(u.CreatedAt),
	}
}

// @Summary		Get user by ID
// @Description	Retrieves a user by their ID
// @Produce		plain
// @Param			id	path		string			true	"User ID"	format(uuid)
// @Success		200	{object}	UserResponse	"User object"
// @Failure		400	{object}	ErrorResponse	"Invalid user ID"
// @Failure		404	{object}	ErrorResponse	"User not found"
// @Router			/users/{id} [get]
func (s *Server) handleGetUser(c fiber.Ctx) error {
	var params struct {
		UserId string `params:"userId"`
	}

	err := c.Bind().URI(&params)
	if err != nil {
		return s.respondWithError(c, fiber.StatusBadRequest, "Invalid request parameters", err)
	}

	userId, err := uuid.Parse(params.UserId)
	if err != nil {
		return s.respondWithError(c, fiber.StatusBadRequest, "Invalid user ID", err)
	}

	user, err := s.svc.Users.GetUser(c.Context(), userId)
	if err != nil {
		return s.respondWithError(c, fiber.StatusNotFound, "User not found", err)
	}

	return c.JSON(toUserResponse(user))
}

// @Summary		Create a new user
// @Description	Creates a new user in the system
// @Produce		json
// @Success		201	{object}	UserResponse	"User object"
// @Failure		500	{object}	ErrorResponse	"Failed to create user"
// @Router			/users [post]
func (s *Server) handleCreateUser(c fiber.Ctx) error {
	user, err := s.svc.Users.CreateUser(c.Context())
	if err != nil {
		return s.respondWithError(c, fiber.StatusInternalServerError, "Failed to create user", err)
	}

	return c.Status(fiber.StatusCreated).JSON(toUserResponse(user))
}
