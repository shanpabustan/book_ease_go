package routes

import "github.com/gofiber/fiber/v2"

func AppRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Golang World!")
	})

	// CREATE YOUR ENDPOINTS HERE

	// --------------------------
}
