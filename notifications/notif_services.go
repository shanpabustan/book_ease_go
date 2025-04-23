package notifications

import (
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"
	"book_ease_go/middleware"
	"book_ease_go/model"
)

func SendNotification(templateName string, userID int, data map[string]interface{}) error {
	tmpl, exists := NotificationTemplates[templateName]
	if !exists {
		return fmt.Errorf("template not found: %s", templateName)
	}

	message, err := ExecuteTemplate(tmpl, data)
	if err != nil {
		return err
	}

	notification := model.Notification{
		UserID:  fmt.Sprintf("%d", userID),
		Message: message,
	}

	if err := middleware.DBConn.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	return nil
}

func ExecuteTemplate(tmpl *template.Template, data map[string]interface{}) (string, error) {
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}
	return sb.String(), nil
}

func CheckOverdueBooks() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			var borrowedBooks []model.BorrowedBook
			if err := middleware.DBConn.Preload("Book").Where("status = ? AND due_date < ?", "Approved", time.Now()).Find(&borrowedBooks).Error; err != nil {
				log.Println("Error fetching overdue books:", err)
				continue
			}
			for _, b := range borrowedBooks {
				data := map[string]interface{}{
					"BookTitle": b.Book.Title,
					"DueDate":   b.DueDate.Format("Jan 02, 2006"),
				}
				if err := SendNotificationTemplate("OverdueBook", b.UserID, data); err != nil {
					log.Println("Error sending overdue book notification:", err)
				}
			}
		}
	}()
}

func CheckPendingReservations() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			var reservations []model.Reservation
			if err := middleware.DBConn.Preload("Book").Where("status = ? AND preferred_pickup_date <= ?", "Pending", time.Now()).Find(&reservations).Error; err != nil {
				log.Println("Error fetching pending reservations:", err)
				continue
			}
			for _, r := range reservations {
				data := map[string]interface{}{
					"BookTitle":           r.Book.Title,
					"PreferredPickupDate": r.PreferredPickupDate.Format("Jan 02, 2006"),
				}
				if err := SendNotificationTemplate("ReservationReminder", r.UserID, data); err != nil {
					log.Println("Error sending reservation reminder:", err)
				}
			}
		}
	}()
}

func CheckExpiredReservations() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			var reservations []model.Reservation
			if err := middleware.DBConn.Preload("Book").Where("status = ? AND expiry < ?", "Pending", time.Now()).Find(&reservations).Error; err != nil {
				log.Println("Error fetching expired reservations:", err)
				continue
			}
			for _, r := range reservations {
				data := map[string]interface{}{
					"BookTitle": r.Book.Title,
				}
				if err := SendNotificationTemplate("ReservationExpired", r.UserID, data); err != nil {
					log.Println("Error sending expired reservation notification:", err)
				}
				r.Status = "Expired"
				if err := middleware.DBConn.Save(&r).Error; err != nil {
					log.Println("Error updating reservation to expired:", err)
				}
			}
		}
	}()
}

func CheckBookReturn() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			var borrowedBooks []model.BorrowedBook
			if err := middleware.DBConn.Preload("Book").Where("status = ?", "Returned").Find(&borrowedBooks).Error; err != nil {
				log.Println("Error fetching returned books:", err)
				continue
			}
			for _, b := range borrowedBooks {
				data := map[string]interface{}{
					"BookTitle": b.Book.Title,
				}
				if err := SendNotificationTemplate("BookReturned", b.UserID, data); err != nil {
					log.Println("Error sending returned book notification:", err)
				}
			}
		}
	}()
}

func RunNotificationJobs() {
	log.Println("🚀 Notification services have started and cron jobs are running...")
	CheckOverdueBooks()
	CheckPendingReservations()
	CheckExpiredReservations()
	CheckBookReturn()
}
