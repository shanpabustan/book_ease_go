package controller

import (
	"book_ease_go/notifications"
	"book_ease_go/model"
	"book_ease_go/middleware"
	"github.com/gofiber/fiber/v2"
)

func RunNotificationManually(c *fiber.Ctx) error {
	notifications.RunNotificationJobs()
	return c.JSON(fiber.Map{"message": "Notifications triggered!"})
}

func FetchNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id") // Get the user_id from query parameter

	var notifications []model.Notification
	if err := middleware.DBConn.Where("user_id = ?", userID).Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch notifications"})
	}

	return c.Status(fiber.StatusOK).JSON(notifications) // Return the notifications as JSON
}