package notifications

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"book_ease_go/middleware"
	"book_ease_go/model"
)

// ========================= HELPERS =========================

// SendNotification sends a notification to a user
// SendNotification sends both in-app and email notifications to a user
func SendNotification(db *gorm.DB, userID, message string) error {
	// Get user details for email
	var user model.User
	if err := db.First(&user, "user_id = ?", userID).Error; err != nil {
		log.Printf("Error fetching user %s: %v", userID, err)
		return err
	}

	// Send in-app notification
	var existing model.Notification
	err := db.Where("user_id = ? AND message = ?", userID, message).First(&existing).Error

	if err == nil {
		// Notification with same message already exists for the user, don't insert duplicate
		log.Printf("Duplicate notification found for user %s: %s", userID, message)
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		// Unexpected error
		log.Printf("Error checking for existing notification: %v", err)
		return err
	}

	// No existing notification, proceed to create
	notification := model.Notification{
		UserID:  userID,
		Message: message,
	}

	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Error creating notification: %v", err)
		return err
	}

	// Send email notification
	subject := "Book Ease Notification"
	htmlBody := fmt.Sprintf(`
		<div style="font-family: 'Poppins', Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8f9fa;">
			<div style="background-color: #008080; padding: 20px; border-radius: 8px 8px 0 0;">
				<h2 style="color: #ffffff; margin: 0; font-weight: 600; font-size: 24px;">Book Ease Notification</h2>
			</div>
			<div style="background-color: #ffffff; padding: 30px; border-radius: 0 0 8px 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<p style="color: #333333; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">%s</p>
				<hr style="border: none; border-top: 1px solid #e0e0e0; margin: 20px 0;">
				<p style="color: #666666; font-size: 12px; margin: 0;">This is an automated message from Book Ease Library Management System.</p>
			</div>
			<div style="text-align: center; margin-top: 20px;">
				<p style="color: #008080; font-size: 12px; margin: 0;">© %d Book Ease Library Management System</p>
			</div>
		</div>
	`, message, time.Now().Year())

	if err := SendEmail(user.Email, subject, htmlBody); err != nil {
		log.Printf("Error sending email to %s: %v", user.Email, err)
		return err
	}

	log.Printf("Successfully sent notification to user %s: %s", userID, message)
	return nil
}

// NotifyAllAdmins notifies all admins with a message
func NotifyAllAdmins(db *gorm.DB, message string) {
	var admins []model.User
	if err := db.Where("user_type = ?", "Admin").Find(&admins).Error; err != nil {
		log.Printf("Error fetching admin users: %v", err)
		return
	}

	log.Printf("Found %d admin users to notify", len(admins))

	for _, admin := range admins {
		if err := SendNotification(db, admin.UserID, message); err != nil {
			log.Printf("Failed to send notification to admin %s: %v", admin.UserID, err)
		} else {
			log.Printf("Successfully sent notification to admin %s", admin.UserID)
		}
	}
}

// ========================= USER NOTIFICATIONS =========================

func NotifyApprovedReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" has been picked up. Enjoy Reading.`, book.Title)
	SendNotification(db, user.UserID, msg)
}

func NotifyPendingReservation(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`Your reservation for "%s" is being reviewed. Please pick up your book at the library counter.`, book.Title)
	SendNotification(db, user.UserID, msg)
}

func NotifyReturnedBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`You have successfully returned "%s". Thank you!`, book.Title)
	SendNotification(db, user.UserID, msg)
}

func NotifyOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`The book "%s" is overdue. Please return it immediately to avoid late penalties.`, book.Title)
	SendNotification(db, user.UserID, msg)
}

func NotifyAccountBlocked(db *gorm.DB, user model.User) {
	msg := `Your account has been temporarily blocked due to multiple overdue books. Please contact the librarian.`
	SendNotification(db, user.UserID, msg)
}

// ========================= ADMIN NOTIFICATIONS =========================

func NotifyAdminReservationRequest(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has requested a reservation for "%s". Please review and take action.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg)
}

func NotifyAdminOverdueBook(db *gorm.DB, user model.User, book model.Book) {
	msg := fmt.Sprintf(`%s %s has not returned "%s", which is now overdue.`, user.FirstName, user.LastName, book.Title)
	NotifyAllAdmins(db, msg)
}

func NotifyAdminNewUser(db *gorm.DB, user model.User) {
	msg := fmt.Sprintf(`A new user, %s %s, has registered.`, user.FirstName, user.LastName)
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

		// Use the centralized MarkOverdueBooks function from middleware
		middleware.MarkOverdueBooks()

		// Fetch newly marked overdue books to send notifications
		var overdue []model.BorrowedBook
		if err := db.Where("due_date < ? AND status = ? AND return_date IS NULL", time.Now(), "Overdue").
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
