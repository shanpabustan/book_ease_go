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
func SendNotification(db *gorm.DB, userID, message string) error {
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
}

// NotifyAllAdmins notifies all admins with a message
func NotifyAllAdmins(db *gorm.DB, message string) {
	var admins []model.User
	if err := db.Where("user_type = ?", "Admin").Find(&admins).Error; err == nil {
		for _, admin := range admins {
			SendNotification(db, admin.UserID, message) // Adjusted to SendNotification
		}
	}
}

// ========================= USER NOTIFICATIONS =========================

func NotifyApprovedReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" has been approved. Ready for pickup at the library counter.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
}

func NotifyPendingReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" is being reviewed. You will be notified once it's approved.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
}

func NotifyReturnedBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`You have successfully returned "%s". Thank you!`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
}

func NotifyOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`The book "%s" is overdue. Please return it immediately to avoid late penalties.`, book.Title)
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
}

func NotifyAccountBlocked(db *gorm.DB, user model.User) {
	msg := `Your account has been temporarily blocked due to multiple overdue books. Please contact the librarian.`
	SendNotification(db, user.UserID, msg) // Adjusted to SendNotification
}

// ========================= ADMIN NOTIFICATIONS =========================

func NotifyAdminReservationRequest(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has requested a reservation for "%s". Please review and take action.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
}

func NotifyAdminOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has not returned "%s", which is now overdue.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
}

func NotifyAdminNewUser(db *gorm.DB, user model.User) {
	msg := fmt.Sprintf(`A new user, %s %s, has registered.`, user.FirstName, user.LastName)
	NotifyAllAdmins(db, msg) // Adjusted to NotifyAllAdmins
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
