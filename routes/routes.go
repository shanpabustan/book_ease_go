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

	//Login User
	login := app.Group("/login")
	login.Post("/login", controller.Login)


	//student routes
	stud := app.Group("/stud")
	stud.Post("/register", controller.CreateStudent)
	stud.Get("/get-all", controller.FetchAllBooks)
	stud.Get("/get-borrowed", controller.FetchBorrowedBooks)


	//Reserve Book
	reserve := app.Group("/reserve")
	reserve.Post("/reserve-book", controller.ReserveBook)
	

	//Admin Routes
	admin :=app.Group("/admin")
	admin.Post("/add-book", controller.AddBook)

	




	
}
