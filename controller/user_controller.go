package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	response "book_ease_go/responses"
	"time"
	"github.com/gofiber/fiber/v2"
)

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

	// Generate a unique ID for reservation_id
	reservation.ReservationID = uint(time.Now().UnixNano())

	// Check if reservation already exists based on reservation_id
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




