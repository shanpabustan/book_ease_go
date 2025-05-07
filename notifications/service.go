package notifications

import (
	"bytes"
	"fmt"
	"book_ease_go/middleware"
	"book_ease_go/model"
)

// SendNotificationTemplate sends a notification only if it hasn’t already been sent
func SendNotificationTemplate(category string, userID string, data map[string]interface{}) error {
	tmpl, exists := NotificationTemplates[category]
	if !exists {
		return fmt.Errorf("notification category not found: %s", category)
	}

	var msg bytes.Buffer
	if err := tmpl.Execute(&msg, data); err != nil {
		return err
	}

	message := msg.String()

	// Check if the same notification message already exists for this user
	var existing model.Notification
	err := middleware.DBConn.Where("user_id = ? AND message = ?", userID, message).First(&existing).Error
	if err == nil {
		// Notification already exists
		return nil
	}

	// Send notification
	notification := model.Notification{
		UserID:  userID,
		Message: message,
	}
	return middleware.DBConn.Create(&notification).Error
}
