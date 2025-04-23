// notifications/service.go
package notifications

import (
	"bytes"
	"fmt"
	"book_ease_go/middleware"
	"book_ease_go/model"
)

// SendNotificationTemplate dynamically renders a message based on category and model-based data
func SendNotificationTemplate(category string, userID string, data map[string]interface{}) error {
	tmpl, exists := NotificationTemplates[category]
	if !exists {
		return fmt.Errorf("notification category not found: %s", category)
	}

	var msg bytes.Buffer
	if err := tmpl.Execute(&msg, data); err != nil {
		return err
	}

	notification := model.Notification{
		UserID:  userID,
		Message: msg.String(),
	}

	return middleware.DBConn.Create(&notification).Error
}
