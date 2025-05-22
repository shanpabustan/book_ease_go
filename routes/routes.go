package routes

import (
	"book_ease_go/controller"

	"github.com/gofiber/fiber/v2"
)

func AppRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Golang World!")
	})

	// CREATE YOUR ENDPOINTS HERE

	// Change password endpoint (for both admin and student)
	app.Post("/change-password", controller.ChangePassword)

	//student routes
	stud := app.Group("/stud")
	stud.Post("/register", controller.CreateStudent)
	stud.Get("/get-borrowed", controller.FetchBorrowedBooks)
	stud.Post("/login", controller.LoginUser)
	stud.Post("/logout", controller.LogOutUser)
	stud.Put("/edit", controller.EditUser)
	stud.Post("/add-pic", controller.UpdateAvatar)
	stud.Get("/get-books-status", controller.FetchBorrowedBooksByStatus)
	stud.Patch("/reset-password", controller.ResetPassword)
	stud.Get("/get-recommended", controller.FetchRecommendedBooks)
	stud.Get("/popular-books", controller.FetchMostPopularBooks)

	//Reserve Book - Student
	reserve := app.Group("/reserve")
	reserve.Post("/reserve-book", controller.ReserveBook)
	//reserve.Post("/cancel-reservation", controller.CancelReservation)

	//admin and student fetch of books.
	app.Get("/get-all", controller.FetchAllBooks)

	//Admin Routes
	admin := app.Group("/admin")
	admin.Post("/add-book", controller.AddBook)
	admin.Put("/edit-book/:book_id", controller.UpdateBook)
	admin.Get("/get-users", controller.GetUsers)
	admin.Get("/count", controller.CountStudents)
	admin.Get("/count-borrowed-books", controller.CountBorrowedBooks)
	admin.Get("/count-reservations", controller.CountReservations)
	admin.Get("/count-overdue-books", controller.CountOverdueBooks)
	admin.Get("/most-borrowed-books", controller.GetMostBorrowedBooks)
	admin.Get("/most-borrowed-categories", controller.GetMostBorrowedCategories)
	
	admin.Put("/disable-students", controller.DisableAllStudents)
	admin.Put("/approve-reservation/:reservation_id", controller.ApproveReservation)
	admin.Put("/cancel-reservation/:reservation_id", controller.DisapproveReservation)
	admin.Put("/return-book/:borrowed_id", controller.ReturnBook)
	admin.Get("/get-reservations", controller.GetAllReservations)
	admin.Get("/export-books", controller.ExportBooks)
	admin.Get("/export-users", controller.ExportUsers)
	admin.Get("/semester/end-date", controller.GetSemesterEndDate)
	admin.Put("/semester/end-date", controller.UpdateSemesterEndDate)
	admin.Post("/semester/auto-disable-students", controller.EndOfSemester)
	admin.Post("/edit-admin", controller.EditAdminUser)
	admin.Get("/get-borrowed-books", controller.GetAllBorrowedBooks)
	admin.Put("/unblock-student/:userID", controller.UnblockUser)
	//admin.Get("/test-email", controller.TestEmailSending)

	app.Get("/test/fetch-notifs", controller.FetchNotifications)
	app.Get("/notifications/unread", controller.FetchUnreadNotifications)
	app.Put("/notifications/mark-as-read/:notification_id", controller.MarkNotificationAsRead)
	app.Get("/test/admin-notification", controller.TestAdminNotification)

}
