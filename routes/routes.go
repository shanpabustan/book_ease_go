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


	//student routes
	stud := app.Group("/stud")
	stud.Post("/register", controller.CreateStudent)
	stud.Get("/get-borrowed", controller.FetchBorrowedBooks)
	stud.Post("/login", controller.LoginUser)
	stud.Post("/logout", controller.LogOutUser)
	stud.Put("/edit", controller.EditUser)
	stud.Post("/add-pic", controller.UpdateAvatar)


	//Reserve Book - Student
	reserve := app.Group("/reserve")
	reserve.Post("/reserve-book", controller.ReserveBook)
	
	//admin and student fetch of books.
	app.Get("/get-all", controller.FetchAllBooks)
	

	//Admin Routes
	admin :=app.Group("/admin")
	admin.Post("/add-book", controller.AddBook)	
	admin.Put("/edit-book/:book_id", controller.UpdateBook)
	admin.Get("/get-users", controller.GetUsers)
	admin.Get("/count", controller.CountStudents)
	admin.Put("/disable-students", controller.DisableAllStudents)
	admin.Put("/approve-reservation/:reservation_id", controller.ApproveReservation)
	admin.Put("/cancel-reservation/:reservation_id", controller.DisapproveReservation)
	admin.Put("/return-book/:borrowed_id", controller.ReturnBook)
	admin.Get("/get-reservations", controller.GetAllReservations)
	admin.Get("/check-penalty/:userID", controller.CheckAndBlockUser)
	admin.Get("/export-books", controller.ExportBooks)
	admin.Get("/export-users", controller.ExportUsers)
	admin.Get("/semester/end-date", controller.GetSemesterEndDate)
	admin.Put("/semester/end-date", controller.UpdateSemesterEndDate)
	admin.Post("/semester/auto-disable-students", controller.EndOfSemester)

	
	

	




	
}
