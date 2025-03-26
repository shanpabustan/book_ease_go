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

	//Register User
	reg := app.Group("/reg")
	reg.Post("/register", controller.CreateStudent)

	//Reserve Book
	reserve := app.Group("/reserve")
	reserve.Post("/reserve-book", controller.ReserveBook)
	

	//Admin Routes
	admin :=app.Group("/admin")
	admin.Post("/add-book", controller.AddBook)

	




	// --------------------------
}
