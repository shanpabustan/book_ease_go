package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	response "book_ease_go/responses"
	"time"
	"github.com/gofiber/fiber/v2"
)

//Register User
func CreateStudent(c *fiber.Ctx) error {
	var student model.User
	if err := c.BodyParser(&student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Check if user already exists based on user_id
	var existingUser model.User
	if err := middleware.DBConn.Debug().Table("users").Where("user_id = ?", student.UserID).First(&existingUser).Error; err == nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "User already exists",
			Data:    nil,
		})
	}

	if err := middleware.DBConn.Debug().Table("users").Create(&student).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Create",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Successful Creation",
		Data:    student,
	})
}


//Book Reservation
func ReserveBook(c *fiber.Ctx) error {
	var reservation model.Reservation
	if err := c.BodyParser(&reservation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// unique id for reservation
	reservation.ReservationID = uint(time.Now().UnixNano())

	// checking of reservation
	var existingReservation model.Reservation
	if err := middleware.DBConn.Debug().Table("reservations").Where("reservation_id = ?", reservation.ReservationID).First(&existingReservation).Error; err == nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Reservation already exists",
			Data:    nil,
		})
	}

	if err := middleware.DBConn.Debug().Table("reservations").Create(&reservation).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Create",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Reservation Successful",
		Data:    reservation,
	})
}

//FETCHING SECTION

// will base on the most borrowed books by the users
func FetchRecommendedBooks(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	var books []model.Book
	bookfetch := middleware.DBConn.Debug().Table("books").
		Where("is_recommended = ? AND user_id = ?", true, userID)

	if err := bookfetch.Find(&books).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Fetch Recommended Books",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Recommended Books Fetched Successfully",
		Data:    books,
	})
}

//fetch borrowed books
func FetchBorrowedBooks(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "User ID is required",
			Data:    nil,
		})
	}

	var books []model.Book
	bookfetch := middleware.DBConn.Debug().Table("books").
		Joins("JOIN reservations ON books.book_id = reservations.book_id").
		Where("reservations.user_id = ? AND reservations.status = ?", userID, "Approved")

	if err := bookfetch.Find(&books).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Fetch Borrowed Books",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Borrowed Books Fetched Successfully",
		Data:    books,
	})
}


func FetchAllBooks(c *fiber.Ctx) error {
	var books []model.Book
	bookfetch := middleware.DBConn.Debug().Table("books")

	if err := bookfetch.Find(&books).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Fetch All Books",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "All Books Fetched Successfully",
		Data:    books,
	})
}




