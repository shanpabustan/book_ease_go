package middleware

import (
	"log"
	"time"

	"book_ease_go/model"
)

// MarkOverdueBooks updates the status of borrowed books past due date.
func MarkOverdueBooks() {
	now := time.Now()
	result := DBConn.Model(&model.BorrowedBook{}).
		Where("due_date < ? AND return_date IS NULL AND status NOT IN ?", now, []string{"Returned", "Overdue", "Damaged"}).
		Update("status", "Overdue")

	if result.Error != nil {
		log.Printf("❌ Failed to mark overdue books: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("📘 Marked %d book(s) as overdue.", result.RowsAffected)
	}
}

// StartOverdueScheduler runs the checker every hour.
func StartOverdueScheduler() {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			MarkOverdueBooks()
		}
	}()
	log.Println("📅 Overdue scheduler started.")
}
