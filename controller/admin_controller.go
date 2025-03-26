package controller


import (
	"book_ease_go/model"
	response "book_ease_go/responses"
	"book_ease_go/middleware"
	"github.com/gofiber/fiber/v2"
)

func AddBook(c *fiber.Ctx) error {
	var book model.Book
	if err := c.BodyParser(&book); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	var existingBook model.Book
	if err := middleware.DBConn.Debug().Table("book").Where("title = ?", book.Title).First(&existingBook).Error; err == nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Book already exists",
			Data:    nil,
		})
	}

	if err := middleware.DBConn.Debug().Table("books").Create(&book).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Add Book",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Book Added Successfully",
		Data:    book,
	})
}