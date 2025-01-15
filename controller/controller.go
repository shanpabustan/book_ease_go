package controller

import (
	"github.com/gofiber/fiber/v2"
)

// SampleController is an example endpoint which returns a
// simple string message.
func SampleController(c *fiber.Ctx) error {
	return c.SendString("Hello, Golang World!")
}
