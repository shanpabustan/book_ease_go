package notifications

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"book_ease_go/model"
	"book_ease_go/middleware"
	
)

// ========================= HELPERS =========================

// SendNotification sends a notification to a user
// SendNotification sends both in-app and email notifications to a user
func SendNotification(db *gorm.DB, userID, message string) error {
	// Get user details for email
	var user model.User
	if err := db.First(&user, "user_id = ?", userID).Error; err != nil {
		return err
	}

	// Send in-app notification
	var existing model.Notification
	err := db.Where("user_id = ? AND message = ?", userID, message).First(&existing).Error

	if err == nil {
		// Notification with same message already exists for the user, don't insert duplicate
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		// Unexpected error
		return err
	}

	// No existing notification, proceed to create
	notification := model.Notification{
		UserID:  userID,
		Message: message,
	}
	return db.Create(&notification).Error
	if err := db.Create(&notification).Error; err != nil {
		return err
	}

	// Send email notification
	subject := "Book Ease Notification"
	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">Book Ease Notification</h2>
			<p style="color: #666;">%s</p>
			<hr style="border: 1px solid #eee;">
			<p style="color: #999; font-size: 12px;">This is an automated message from Book Ease Library Management System.</p>
		</div>
	`, message)

	return SendEmail(user.Email, subject, htmlBody)
}

// NotifyAllAdmins notifies all admins with a message
func NotifyAllAdmins(db *gorm.DB, message string) {
	var admins []model.User
	if err := db.Where("user_type = ?", "Admin").Find(&admins).Error; err == nil {
		for _, admin := range admins {
			SendNotification(db, admin.UserID, message) // Adjusted to SendNotification
			SendNotification(db, admin.UserID, message)
		}
	}
}

// ========================= USER NOTIFICATIONS =========================

func NotifyApprovedReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" has been approved. Ready for pickup at the library counter.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
	SendNotification(db, user.UserID, msg)
}

func NotifyPendingReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" is being reviewed. You will be notified once it's approved.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
	SendNotification(db, user.UserID, msg)
}

func NotifyReturnedBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`You have successfully returned "%s". Thank you!`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
	SendNotification(db, user.UserID, msg)
}

func NotifyOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`The book "%s" is overdue. Please return it immediately to avoid late penalties.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
	SendNotification(db, user.UserID, msg)
}

func NotifyAccountBlocked(db *gorm.DB, user model.User) {
	msg := `Your account has been temporarily blocked due to multiple overdue books. Please contact the librarian.`
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
	SendNotification(db, user.UserID, msg)
}

// ========================= ADMIN NOTIFICATIONS =========================

func NotifyAdminReservationRequest(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has requested a reservation for "%s". Please review and take action.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
	NotifyAllAdmins(db, msg)
}

func NotifyAdminOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has not returned "%s", which is now overdue.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
	NotifyAllAdmins(db, msg)
}

func NotifyAdminNewUser(db *gorm.DB, user model.User) {
	msg := fmt.Sprintf(`A new user, %s %s, has registered.`, user.FirstName, user.LastName)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
	NotifyAllAdmins(db, msg)
}

// ========================= CRON JOB =========================

// StartOverdueCheckerCron schedules a job that runs every midnight
func StartOverdueCheckerCron() {
	c := cron.New()

	// Every 30 seconds
	_, err := c.AddFunc("@every 30s", func() {
		log.Println("[CRON] Running overdue checker...")
		db := middleware.DBConn

		var overdue []model.BorrowedBook
		if err := db.Where("due_date < ? AND status != ?", time.Now(), "Returned").
			Find(&overdue).Error; err != nil {
			log.Printf("Error fetching overdue books: %v", err)
			return
		}

		for _, entry := range overdue {
			var user model.User
			var book model.Book
			db.First(&user, "user_id = ?", entry.UserID)
			db.First(&book, "book_id = ?", entry.BookID)

			NotifyOverdueBook(db, user, book)
			NotifyAdminOverdueBook(db, user, book)
		}
	})

	if err != nil {
		log.Fatalf("Failed to schedule overdue checker cron job: %v", err)
	}

	c.Start()
	log.Println("✅ Overdue checker cron job started (every 30s).")
}