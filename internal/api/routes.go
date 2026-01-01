package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

func handleHealth(c fiber.Ctx) error {
	return c.SendString("OK")
}

func handleGetUser(c fiber.Ctx) error {
	id := c.Params("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Invalid UUID format",
		})
	}

	return c.JSON(UserResponse{
		ID:        uid,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	})
}

func handleCreateUser(c fiber.Ctx) error {
	id, err := uuid.NewV7()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Message: "Failed to generate UUID",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:        id,
		CreatedAt: time.Now(),
	})
}
