package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	"book_ease_go/notifications"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// RunNotificationManually will trigger the sending of notifications
func RunNotificationManually(c *fiber.Ctx) error {
	// Example: Send a notification to a specific user
	userID := c.Query("user_id")
	bookID := c.Query("book_id")
	message := c.Query("message")

	if userID == "" || bookID == "" || message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required query parameters (user_id, book_id, message)",
		})
	}

	// Retrieve user and book details from the database
	var user model.User
	var book model.Book
	if err := middleware.DBConn.First(&user, "user_id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	if err := middleware.DBConn.First(&book, "book_id = ?", bookID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
	}

	// Send notification manually
	err := notifications.SendNotification(middleware.DBConn, userID, message) // Capitalize SendNotification
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send notification"})
	}

	// Return a success message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Notification sent successfully!",
	})
}

// FetchNotifications will retrieve notifications for a specific user
func FetchNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id") // Get the user_id from query parameter

	var notifications []model.Notification
	if err := middleware.DBConn.Where("user_id = ?", userID).Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch notifications"})
	}

	return c.Status(fiber.StatusOK).JSON(notifications) // Return the notifications as JSON
}

// FetchUnreadNotifications returns only unread notifications for a user
func FetchUnreadNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing user_id query parameter",
		})
	}

	var unreadNotifs []model.Notification
	if err := middleware.DBConn.
		Where("user_id = ? AND is_read = false", userID).
		Order("created_at DESC").
		Find(&unreadNotifs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to fetch unread notifications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(unreadNotifs)
}

// MarkNotificationAsRead marks a single notification as read by ID
func MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationID := c.Params("notification_id")

	var notif model.Notification
	if err := middleware.DBConn.First(&notif, notificationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	notif.IsRead = true
	if err := middleware.DBConn.Save(&notif).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update notification",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Notification marked as read",
	})
}

// TestAdminNotification will test sending a notification to all admins
func TestAdminNotification(c *fiber.Ctx) error {
	var admins []model.User
	if err := middleware.DBConn.Debug().Where("user_type = ?", "Admin").Find(&admins).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to find admins: %v", err),
		})
	}

	// Print found admins
	log.Printf("Found %d admins", len(admins))
	for _, admin := range admins {
		log.Printf("Admin: ID=%s, Email=%s", admin.UserID, admin.Email)
	}

	// Try to send a test notification
	testMsg := "This is a test notification for admins"
	notifications.NotifyAllAdmins(middleware.DBConn, testMsg)

	return c.JSON(fiber.Map{
		"message":       "Test notification attempted",
		"admins_found":  len(admins),
		"admin_details": admins,
	})
}
