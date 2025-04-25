package controller

import (
	"book_ease_go/exports"
	"book_ease_go/middleware"
	"book_ease_go/model"
	"book_ease_go/notifications"
	response "book_ease_go/responses"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

//Data Analytics
//Overall Registered Users "Student
func CountStudents(c *fiber.Ctx) error {
	var count int64

	// Count the number of users with user_type = "Student"
	if err := middleware.DBConn.Debug().
		Table("users").
		Where("user_type = ?", "Student").
		Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to count students",
			Data:    err.Error(),
		})
	}

	// Return the count of students
	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Student count fetched successfully",
		Data:    count,
	})
}

func EditAdminUser(c *fiber.Ctx) error {
	var request struct {
		UserID   string `json:"user_id"`
		Password string `json:"password"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "Invalid request payload",
			"Error":   err.Error(),
		})
	}

	// Find user by ID, ensuring the user type is "Admin"
	var user model.User
	if err := middleware.DBConn.Table("users").Where("user_id = ? AND user_type = ?", request.UserID, "Admin").First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"RetCode": "404",
			"Message": "Admin user not found",
		})
	}

	// Hash the new password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to hash the password",
			"Error":   err.Error(),
		})
	}

	// Update the admin user's password with the hashed version
	user.Password = string(hashedPassword)

	// Save changes
	if err := middleware.DBConn.Table("users").Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to update admin user",
			"Error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"RetCode": "200",
		"Message": "Admin user password updated successfully",
		"Data":    user,
	})
}



// Add or Register a Book
func AddBook(c *fiber.Ctx) error {
	var book model.Book

	// Parse body
	if err := c.BodyParser(&book); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Check if the book already exists by title or ISBN
	var existingBook model.Book
	if err := middleware.DBConn.
		Table("books").
		Where("title = ? OR isbn = ?", book.Title, book.ISBN).
		First(&existingBook).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "Book already exists",
			Data:    nil,
		})
	}

	// Auto-set AvailableCopies to match TotalCopies
	book.AvailableCopies = book.TotalCopies

	// Save to DB
	if err := middleware.DBConn.Create(&book).Error; err != nil {
		fmt.Println("Error inserting book:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to add book",
			Data:    err.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Book added successfully",
		Data:    book,
	})
}



// Update Details of the book
func UpdateBook(c *fiber.Ctx) error {
	var book model.Book

	// Parse the request body to get updated book details
	if err := c.BodyParser(&book); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Check if the book exists based on book_id
	var existingBook model.Book
	if err := middleware.DBConn.Table("books").Where("book_id = ?", book.BookID).First(&existingBook).Error; err != nil {
		// Return an error if the book is not found
		if err.Error() == "record not found" {
			return c.JSON(response.ResponseModel{
				RetCode: "404",
				Message: "Book not found",
				Data:    nil,
			})
		}
		// Return other database errors
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch book",
			Data:    err.Error(),
		})
	}

	// Update the fields of the existing book
	// You can update only the fields that have been changed, keeping it efficient
	if err := middleware.DBConn.Table("books").Where("book_id = ?", book.BookID).Updates(model.Book{
		Title:           book.Title,
		Author:          book.Author,
		Category:        book.Category,
		ISBN:            book.ISBN,
		LibrarySection:  book.LibrarySection,
		ShelfLocation:   book.ShelfLocation,
		TotalCopies:     book.TotalCopies,
		AvailableCopies: book.AvailableCopies,
		BookCondition:   book.BookCondition,
		Picture:         book.Picture,
	}).Error; err != nil {
		// Log and return the error
		fmt.Printf("Error updating book: %v\n", err)
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update book",
			Data:    err.Error(),
		})
	}

	// Return success response
	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Book updated successfully",
		Data:    book,
	})
}


// Get Students
func GetUsers(c *fiber.Ctx) error {
    var users []model.User

    // Fetch only users with user_type = "Student"
    if err := middleware.DBConn.Debug().
        Table("users").
        Where("user_type = ?", "Student").
        Find(&users).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to fetch users",
            Data:    err.Error(),
        })
    }

    // Format the user data
    var formattedUsers []map[string]interface{}
    for _, user := range users {
        // Handle optional fields inline
        middleName := ""
        if user.MiddleName != nil {
            middleName = *user.MiddleName
        }

        suffix := ""
        if user.Suffix != nil {
            suffix = *user.Suffix
        }

        yearLevel := ""
        if user.YearLevel != nil {
            yearLevel = *user.YearLevel
        }

        program := ""
        if user.Program != nil {
            program = *user.Program
        }

        // Build full name with clean formatting
        fullName := fmt.Sprintf("%s, %s", user.LastName, user.FirstName)
        if middleName != "" {
            fullName += " " + middleName
        }
        if suffix != "" {
            fullName += " " + suffix
        }

        formattedUsers = append(formattedUsers, map[string]interface{}{
            "user_id":    user.UserID,
            "name":       fullName,
            "email":     user.Email,  
            "program":    program,     
            "year_level": yearLevel,
            "contact_number": user.ContactNumber,
            "is_active":      user.IsActive,
            
        })
    }

    return c.JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Users fetched successfully",
        Data:    formattedUsers,
    })
}


//Disable Student Accounts
func DisableAllStudents(c *fiber.Ctx) error {
	result := middleware.DBConn.
		Model(&model.User{}).
		Where("user_type = ? AND is_active = ?", "Student", true).
		Update("is_active", false)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to disable students",
			Data:    result.Error.Error(),
		})
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: fmt.Sprintf("Successfully disabled %d students", result.RowsAffected),
		Data:    nil,
	})
}

// Admin - Approve reservation and create borrowed book record (Picked Up remarks).
// ApproveReservation function with notifications
func ApproveReservation(c *fiber.Ctx) error {
    reservationIDStr := c.Params("reservation_id")
    reservationID, err := strconv.Atoi(reservationIDStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Invalid reservation ID",
            Data:    err.Error(),
        })
    }

    var reservation model.Reservation
    if err := middleware.DBConn.First(&reservation, "reservation_id = ?", reservationID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
            RetCode: "404",
            Message: "Reservation not found",
            Data:    nil,
        })
    }

    if reservation.Status != "Pending" {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Only pending reservations can be approved",
            Data:    nil,
        })
    }

    tx := middleware.DBConn.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    var book model.Book
    if err := tx.First(&book, reservation.BookID).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Book not found",
            Data:    err.Error(),
        })
    }

    if book.AvailableCopies <= 0 {
        tx.Rollback()
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "No available copies for this book",
            Data:    nil,
        })
    }

    book.AvailableCopies--
    if err := tx.Save(&book).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to update book availability",
            Data:    err.Error(),
        })
    }

    reservation.Status = "Approved"
    if err := tx.Save(&reservation).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to approve reservation",
            Data:    err.Error(),
        })
    }

    borrowed := model.BorrowedBook{
        ReservationID:       reservation.ReservationID,
        UserID:              reservation.UserID,
        BookID:              reservation.BookID,
        BorrowDate:          time.Now(),
        DueDate:             time.Now().AddDate(0, 0, 7),
        Status:              "Approved",
        BookConditionBefore: book.BookCondition,
    }

    if err := tx.Create(&borrowed).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Reservation approved, but failed to create borrowed book entry",
            Data:    err.Error(),
        })
    }

    if err := tx.Commit().Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to commit transaction",
            Data:    err.Error(),
        })
    }

    go func() {
    err := notifications.SendNotificationTemplate("ReservationApproved", reservation.UserID, map[string]interface{}{
        "BookTitle": book.Title,
    })
    if err != nil {
		log.Printf("Notification error (approval): %v | UserID: %s | ReservationID: %d | Payload: %+v\n",
			err, reservation.UserID, reservation.ReservationID, map[string]any{
				"BookTitle": book.Title,
			})
    }
}()

    return c.JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Reservation approved and book borrowed successfully",
        Data:    borrowed,
    })
}

// DisapproveReservation function with notifications
func DisapproveReservation(c *fiber.Ctx) error {
    reservationID := c.Params("reservation_id")

    var reservation model.Reservation
    if err := middleware.DBConn.First(&reservation, "reservation_id = ?", reservationID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
            RetCode: "404",
            Message: "Reservation not found",
            Data:    nil,
        })
    }

    if reservation.Status != "Pending" {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Only pending reservations can be disapproved",
            Data:    nil,
        })
    }

    reservation.Status = "Cancelled"
    if err := middleware.DBConn.Save(&reservation).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to disapprove reservation",
            Data:    err.Error(),
        })
    }

    var book model.Book
    _ = middleware.DBConn.First(&book, reservation.BookID).Error

    go func() {
        if err := notifications.SendNotificationTemplate("ReservationDeclined", reservation.UserID, map[string]interface{}{
            "BookTitle": book.Title,
        }); err != nil {
            log.Println("Notification error (decline):", err)
        }
    }()

    return c.JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Reservation disapproved successfully",
        Data:    reservation,
    })
}



func ReturnBook(c *fiber.Ctx) error {
    borrowedIDStr := c.Params("borrowed_id")
    borrowedID, err := strconv.Atoi(borrowedIDStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Invalid borrowed book ID",
            Data:    err.Error(),
        })
    }

    // Get return details from admin input
    var reqData struct {
        BookConditionAfter string  `json:"book_condition_after"`
        PenaltyAmount      float64 `json:"penalty_amount"`
    }
    if err := c.BodyParser(&reqData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Invalid request body",
            Data:    err.Error(),
        })
    }

    tx := middleware.DBConn.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    var borrowed model.BorrowedBook
    if err := tx.First(&borrowed, "borrow_id = ?", borrowedID).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
            RetCode: "404",
            Message: "Borrowed book record not found",
            Data:    nil,
        })
    }

    if borrowed.Status == "Returned" || borrowed.Status == "Overdue" {
        tx.Rollback()
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "This book has already been returned",
            Data:    nil,
        })
    }

    var book model.Book
    if err := tx.First(&book, borrowed.BookID).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Book not found",
            Data:    err.Error(),
        })
    }

    // Increase available copies
    book.AvailableCopies++
    if err := tx.Save(&book).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to update book availability",
            Data:    err.Error(),
        })
    }

    // Set return info
    now := time.Now()
    borrowed.ReturnDate = &now

    // Determine return status
    if now.After(borrowed.DueDate) {
        borrowed.Status = "Overdue"
    } else {
        borrowed.Status = "Returned"
    }

    // Set BookConditionAfter and PenaltyAmount from request
    if reqData.BookConditionAfter != "" {
        borrowed.BookConditionAfter = &reqData.BookConditionAfter
    }

    if reqData.PenaltyAmount > 0 {
        borrowed.PenaltyAmount = reqData.PenaltyAmount
    }

    if err := tx.Save(&borrowed).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to update borrowed book record",
            Data:    err.Error(),
        })
    }

    if err := tx.Commit().Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to commit transaction",
            Data:    err.Error(),
        })
    }

    return c.JSON(response.ResponseModel{
        RetCode: "200",
        Message: "Book return recorded successfully",
        Data:    borrowed,
    })
}




func GetAllReservations(c *fiber.Ctx) error {
	type ReservationResponse struct {
		ReservationID       int       `json:"reservation_id"`
		UserID              string    `json:"user_id"`
		FullName            string    `json:"full_name"`
		BookID              int       `json:"book_id"`
		BookTitle           string    `json:"book_title"`
		PreferredPickupDate time.Time `json:"preferred_pickup_date"`
		Status              string    `json:"status"`
		CreatedAt           time.Time `json:"created_at"`
	}

	var reservations []model.Reservation

	// Ensure preloads are valid: User.UserID is string
	err := middleware.DBConn.
		Preload("User").
		Preload("Book").
		Find(&reservations).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch reservations",
		})
	}

	var response []ReservationResponse

	for _, res := range reservations {
		fullName := formatFullName(
			res.User.LastName,
			res.User.FirstName,
			res.User.MiddleName,
			res.User.Suffix,
		)

		response = append(response, ReservationResponse{
			ReservationID:       res.ReservationID,
			UserID:              res.UserID,
			FullName:            fullName,
			BookID:              res.BookID,
			BookTitle:           res.Book.Title,
			PreferredPickupDate: res.PreferredPickupDate,
			Status:              res.Status,
			CreatedAt:           res.CreatedAt,
		})
	}

	return c.JSON(response)
}

func GetAllBorrowedBooks(c *fiber.Ctx) error {
	type BorrowedBookResponse struct {
		BorrowID            int        `json:"borrow_id"`
		ReservationID       int        `json:"reservation_id"`
		UserID              string     `json:"user_id"`
		FullName            string     `json:"full_name"`
		BookID              int        `json:"book_id"`
		BookTitle           string     `json:"book_title"`
		BorrowDate          time.Time  `json:"borrow_date"`
		DueDate             time.Time  `json:"due_date"`
		ReturnDate          *time.Time `json:"return_date,omitempty"`
		Status              string     `json:"status"`
		BookConditionBefore string     `json:"book_condition_before"`
		BookConditionAfter  *string    `json:"book_condition_after,omitempty"`
		PenaltyAmount       float64    `json:"penalty_amount"`
	}

	type borrowedBookRaw struct {
		BorrowID            int
		ReservationID       int
		UserID              string
		BookID              int
		BorrowDate          time.Time
		DueDate             time.Time
		ReturnDate          *time.Time
		Status              string
		BookConditionBefore string
		BookConditionAfter  *string
		PenaltyAmount       float64
		FirstName           string
		MiddleName          *string
		LastName            string
		Suffix              *string
		BookTitle           string
	}

	var rawData []borrowedBookRaw

	query := `
		SELECT 
			bb.borrow_id,
			bb.reservation_id,
			bb.user_id,
			bb.book_id,
			bb.borrow_date,
			bb.due_date,
			bb.return_date,
			bb.status,
			bb.book_condition_before,
			bb.book_condition_after,
			bb.penalty_amount,
			u.first_name,
			u.middle_name,
			u.last_name,
			u.suffix,
			b.title AS book_title
		FROM borrowed_books bb
		JOIN users u ON bb.user_id = u.user_id
		JOIN books b ON bb.book_id = b.book_id
		ORDER BY bb.borrow_date DESC
	`

	if err := middleware.DBConn.Raw(query).Scan(&rawData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch borrowed books",
		})
	}

	var response []BorrowedBookResponse
	for _, r := range rawData {
		fullName := formatFullName(r.LastName, r.FirstName, r.MiddleName, r.Suffix)

		response = append(response, BorrowedBookResponse{
			BorrowID:            r.BorrowID,
			ReservationID:       r.ReservationID,
			UserID:              r.UserID,
			FullName:            fullName,
			BookID:              r.BookID,
			BookTitle:           r.BookTitle,
			BorrowDate:          r.BorrowDate,
			DueDate:             r.DueDate,
			ReturnDate:          r.ReturnDate,
			Status:              r.Status,
			BookConditionBefore: r.BookConditionBefore,
			BookConditionAfter:  r.BookConditionAfter,
			PenaltyAmount:       r.PenaltyAmount,
		})
	}

	return c.JSON(response)
}



func formatFullName(lastName, firstName string, middleName, suffix *string) string {
	fullName := lastName + ", " + firstName

	if middleName != nil && *middleName != "" {
		fullName += " " + *middleName
	}
	if suffix != nil && *suffix != "" {
		fullName += " " + *suffix
	}
	return fullName
}

func CheckAndBlockUser(c *fiber.Ctx) error {
    userID := c.Params("userID")

    // First, check if user exists
    var user model.User
    if err := middleware.DBConn.First(&user, "user_id = ?", userID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "User not found",
        })
    }

    // Get user's borrowed books, ordered by borrow date descending
    var borrowedBooks []model.BorrowedBook
    if err := middleware.DBConn.
        Where("user_id = ?", userID).
        Order("borrow_date DESC").
        Find(&borrowedBooks).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch borrowed books",
        })
    }

    // Count 3 consecutive penalties
    penaltyCount := 0
    for i, book := range borrowedBooks {
        // Automatically update to "Overdue" if past due date and not returned
        if book.Status != "Returned" && book.ReturnDate == nil && time.Now().After(book.DueDate) {
            borrowedBooks[i].Status = "Overdue"

            // Update status in DB
            err := middleware.DBConn.Model(&model.BorrowedBook{}).
                Where("borrow_id = ?", book.BorrowID).
                Update("status", "Overdue").Error
            if err != nil {
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "error": "Failed to update book status",
                })
            }
        }

        // Recalculate isLate based on updated status
        isLate := borrowedBooks[i].Status == "Overdue" || 
            (book.ReturnDate != nil && book.ReturnDate.After(book.DueDate))
        
        if isLate {
            penaltyCount++
        } else {
            penaltyCount = 0
        }

        if penaltyCount >= 3 {
            if user.IsActive {
                user.IsActive = false
                if err := middleware.DBConn.Save(&user).Error; err != nil {
                    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                        "error": "Failed to update user status",
                    })
                }
            }

            return c.JSON(fiber.Map{
                "message": "User has been blocked due to 3 consecutive penalties.",
                "user_id": userID,
            })
        }
    }

    return c.JSON(fiber.Map{
        "message": "User is still active.",
        "user_id": userID,
    })
}







func ExportBooks(c *fiber.Ctx) error {
	var books []model.Book
	if err := middleware.DBConn.Find(&books).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch books",
			Data:    err.Error(),
		})
	}

	format := c.Query("format", "csv") // default format
	var content []byte
	var contentType, filename string
	var err error

	switch format {
	case "csv":
		content, err = export.ExportBooksCSV(books)
		contentType = "text/csv"
		filename = "books.csv"
	case "excel":
		content, err = export.ExportBooksExcel(books)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "books.xlsx"
	case "pdf":
		content, err = export.ExportBooksPDF(books)
		contentType = "application/pdf"
		filename = "books.pdf"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid format specified. Use 'csv', 'excel', or 'pdf'.",
			Data:    nil,
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: fmt.Sprintf("Failed to export books to %s", format),
			Data:    err.Error(),
		})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(content)
}




func ExportUsers(c *fiber.Ctx) error {
	var users []model.User
	if err := middleware.DBConn.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch users",
			Data:    err.Error(),
		})
	}

	format := c.Query("format", "csv") // default to csv
	var fileContent []byte
	var contentType string
	var fileExtension string
    var err error

	// Export based on the format specified in query params
	switch format {
	case "csv":
		fileContent, err = export.ExportUsersCSV(users)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to export users to CSV",
				Data:    err.Error(),
			})
		}
		contentType = "text/csv"
		fileExtension = "csv"

	case "excel":
		fileContent, err = export.ExportUsersExcel(users)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to export users to Excel",
				Data:    err.Error(),
			})
		}
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		fileExtension = "xlsx"

	case "pdf":
		fileContent, err = export.ExportUsersPDF(users)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to export users to PDF",
				Data:    err.Error(),
			})
		}
		contentType = "application/pdf"
		fileExtension = "pdf"

	default:
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid format specified. Use 'csv', 'excel', or 'pdf'.",
			Data:    nil,
		})
	}

	// Set response headers for file download
	filename := fmt.Sprintf("users_export.%s", fileExtension)
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Send the generated file content as response
	return c.Send(fileContent)
}

func GetSemesterEndDate(c *fiber.Ctx) error {
	var setting model.Setting
	if err := middleware.DBConn.Where("key = ?", "semester_end_date").First(&setting).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Semester end date not set",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Semester end date retrieved successfully",
		Data:    setting,
	})
}


func UpdateSemesterEndDate(c *fiber.Ctx) error {
	type Request struct {
		Value string `json:"value"` // Format: YYYY-MM-DD
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", body.Value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid date format. Use YYYY-MM-DD.",
			Data:    nil,
		})
	}

	setting := model.Setting{Key: "semester_end_date"}
	if err := middleware.DBConn.FirstOrCreate(&setting, model.Setting{Key: "semester_end_date"}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to set semester end date",
			Data:    err.Error(),
		})
	}

	setting.Value = body.Value
	if err := middleware.DBConn.Save(&setting).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update semester end date",
			Data:    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Semester end date updated successfully",
		Data:    setting,
	})
}


func EndOfSemester(c *fiber.Ctx) error {
	var setting model.Setting
	if err := middleware.DBConn.Where("key = ?", "semester_end_date").First(&setting).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Semester end date not set",
			Data:    nil,
		})
	}

	semesterEndDate, err := time.Parse("2006-01-02", setting.Value)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid semester end date format",
			Data:    err.Error(),
		})
	}

	// Only proceed if today is on or after the end of semester
	if time.Now().After(semesterEndDate) || time.Now().Equal(semesterEndDate) {
		// Disable all active students
		if err := middleware.DBConn.Model(&model.User{}).
			Where("user_type = ? AND is_active = ?", "Student", true).
			Update("is_active", false).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
				RetCode: "500",
				Message: "Failed to disable students",
				Data:    err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "Students have been disabled after semester end",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Semester has not ended yet. No students were disabled.",
		Data:    nil,
	})
}




