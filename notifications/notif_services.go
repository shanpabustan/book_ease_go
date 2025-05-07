package notifications

import (
	"log"
	"time"

	"book_ease_go/middleware"
	"book_ease_go/model"
)


func CheckOverdueBooks() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for {
            <-ticker.C
            var borrowedBooks []model.BorrowedBook
            if err := middleware.DBConn.Preload("Book").Where("status = ? AND due_date < ?", "Overdue", time.Now().UTC()).Find(&borrowedBooks).Error; err != nil {
                log.Println("Error fetching overdue books:", err)
                continue
            }
            for _, b := range borrowedBooks {
                notificationData := model.NotificationData{
                    BookTitle: b.Book.Title,
                    DueDate:   b.DueDate.Format("Jan 02, 2006"),
                }

                // Convert NotificationData to map[string]interface{}
                data := map[string]interface{}{
                    "book_title": notificationData.BookTitle,
                    "due_date":   notificationData.DueDate,
                }

                err := SendNotificationTemplate("OverdueBook", b.UserID, data)
                if err != nil {
                    log.Println("Error sending overdue book notification:", err)
                }
            }
        }
    }()
}


func CheckPendingReservations() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for {
            <-ticker.C

            now := time.Now().UTC()
            startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
            endOfDay := startOfDay.Add(24 * time.Hour)

            var reservations []model.Reservation
            if err := middleware.DBConn.Preload("Book").Where("status = ? AND preferred_pickup_date >= ? AND preferred_pickup_date < ?", "Pending", startOfDay, endOfDay).Find(&reservations).Error; err != nil {
                log.Println("Error fetching pending reservations:", err)
                continue
            }
            for _, r := range reservations {
                notificationData := model.NotificationData{
                    BookTitle:           r.Book.Title,
                    PreferredPickupDate: r.PreferredPickupDate.Format("Jan 02, 2006"),
                }

                // Convert NotificationData to map[string]interface{}
                data := map[string]interface{}{
                    "title":           notificationData.BookTitle,
                    "preferred_pickup_date": notificationData.PreferredPickupDate,
                }

                err := SendNotificationTemplate("ReservationReminder", r.UserID, data)
                if err != nil {
                    log.Println("Error sending reservation reminder:", err)
                }
            }
        }
    }()
}




func CheckExpiredReservations() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for {
            <-ticker.C
            var reservations []model.Reservation
            if err := middleware.DBConn.Preload("Book").Where("status = ? AND expired_at < ?", "Pending", time.Now().UTC()).Find(&reservations).Error; err != nil {
                log.Println("Error fetching expired reservations:", err)
                continue
            }
            for _, r := range reservations {
                notificationData := model.NotificationData{
                    BookTitle: r.Book.Title,
                }

                // Convert NotificationData to map[string]interface{}
                data := map[string]interface{}{
                    "book_title": notificationData.BookTitle,
                }

                err := SendNotificationTemplate("ReservationExpired", r.UserID, data)
                if err != nil {
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
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for {
            <-ticker.C
            var borrowedBooks []model.BorrowedBook
            if err := middleware.DBConn.Preload("Book").Where("status = ?", "Returned").Find(&borrowedBooks).Error; err != nil {
                log.Println("Error fetching returned books:", err)
                continue
            }
            for _, b := range borrowedBooks {
                notificationData := model.NotificationData{
                    BookTitle: b.Book.Title,
                }

                // Convert NotificationData to map[string]interface{}
                data := map[string]interface{}{
                    "book_title": notificationData.BookTitle,
                }

                err := SendNotificationTemplate("BookReturned", b.UserID, data)
                if err != nil {
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
