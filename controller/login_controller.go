package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"time"
	"github.com/google/uuid"
)

func Login(c *fiber.Ctx) error {
	var login model.User

	if err := c.BodyParser(&login); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	var users model.User
	if err := middleware.DBConn.Table("users").
		Where("user_id = ?", login.UserID).
		First(&users).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users.PasswordHash), []byte(login.PasswordHash)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":    users.UserID,
			"user_type":  users.UserType,
			"session_id": uuid.New().String(), // Unique session ID
			"exp":        time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		})

		// Sign the token with the secret key
		secretKey := middleware.GetEnv("JWT_SECRET")
		tokenString, err := token.SignedString([]byte(secretKey))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

	// Return token and redirect based on UserType
	switch users.UserType {
	case "Admin":
		return c.JSON(fiber.Map{"redirect": "/admin", "token": tokenString})
	case "Student":
		return c.JSON(fiber.Map{"redirect": "/student", "token": tokenString})
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user type"})
	}
}