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

	// Parse request body
	if err := c.BodyParser(&login); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Find user by ID
	var users model.User
	if err := middleware.DBConn.Table("users").
		Where("user_id = ?", login.UserID).
		First(&users).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Compare hashed password
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

	// Sign the token
	secretKey := middleware.GetEnv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// **Fetch additional user data**
	var reservations []model.Reservation
	middleware.DBConn.Table("reservations").Where("user_id = ?", users.UserID).Find(&reservations)

	var borrowedBooks []model.BorrowedBook
	middleware.DBConn.Table("borrowedbooks").Where("user_id = ?", users.UserID).Find(&borrowedBooks)


	// Return response with token and all user-related data
	return c.JSON(fiber.Map{
		"token":         tokenString,
		"user":          users,
		"reservations":  reservations,
		"borrowedBooks": borrowedBooks,
		"redirect":      getRedirectPath(users.UserType),
	})
}

// Function to determine redirect path
func getRedirectPath(userType string) string {
	switch userType {
	case "Admin":
		return "/admin"
	case "Student":
		return "/student"
	default:
		return "/"
	}
}
