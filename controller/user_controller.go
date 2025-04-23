package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	response "book_ease_go/responses"
	"errors"
	"fmt"

	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register User
func CreateStudent(c *fiber.Ctx) error {
	var student model.User

	// Parse JSON body
	if err := c.BodyParser(&student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "Invalid request data",
			"Error":   err.Error(),
		})
	}

	// ✅ Check if user_id already exists
	var existingUser model.User
	err := middleware.DBConn.Table("users").Where("user_id = ?", student.UserID).First(&existingUser).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("DB Query Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Database Error",
			"Error":   err.Error(),
		})
	}

	// If user exists, return error
	if err == nil {
		fmt.Println("User already exists:", student.UserID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "User already exists",
		})
	}

	// ✅ Force user_type to "Student"
	student.UserType = "Student"
	student.AvatarPath = ""

	// ✅ Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to hash password",
			"Error":   err.Error(),
		})
	}
	
	student.Password = string(hashedPassword)


	// ✅ Insert new user
	if err := middleware.DBConn.Table("users").Create(&student).Error; err != nil {
		fmt.Println("DB Insert Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to Create User",
			"Error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"RetCode": "201",
		"Message": "Registration Successful",
		"Data":    student,
	})
}

func LoginUser(c *fiber.Ctx) error {
	var input model.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    err.Error(),
		})
	}

	var users model.User
	if err := middleware.DBConn.Table("users").Where("user_id = ?", input.UserID).First(&users).Error; err != nil {
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
				RetCode: "401",
				Message: "Invalid user ID or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Server error",
			Data:    err.Error(),
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Invalid user ID or password",
		})
	}

	var redirectURL string
	switch users.UserType {
	case "Admin":
		redirectURL = "/admin/dashboard"
	case "Student":
		redirectURL = "/student/dashboard"
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized user type",
		})
	}

	// 🛠 Enable SQL Debugging
	middleware.DBConn = middleware.DBConn.Debug()

	var borrowedBooksWithDetails []model.BorrowedBookWithDetails
	if err := middleware.DBConn.Table("borrowed_books").
	Select(`borrowed_books.reservation_id, 
	        borrowed_books.user_id, 
	        borrowed_books.book_id, 
	        books.title, 
	        books.picture, 
	        borrowed_books.borrow_date, 
	        borrowed_books.due_date`).
	Joins("JOIN books ON borrowed_books.book_id = books.book_id").
	Where("borrowed_books.user_id = ? AND borrowed_books.status = ?", users.UserID, "Approved").
	Scan(&borrowedBooksWithDetails).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch borrowed books",
			Data:    err.Error(),
		})
	}

	// Ensure borrowedBooksWithDetails is at least an empty slice
	if borrowedBooksWithDetails == nil {
		borrowedBooksWithDetails = []model.BorrowedBookWithDetails{}
	}

		// ✅ Set login cookie
		c.Cookie(&fiber.Cookie{
			Name:     "user_session",
			Value:    users.UserID,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   false, // set to true if using HTTPS
			SameSite: "Lax",
		})


	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Login successful",
		Data: fiber.Map{
			"user_id":        users.UserID,
			"user_type":      users.UserType,
			"last_name":      users.LastName,
			"first_name":     users.FirstName,
			"middle_name":    users.MiddleName,
			"suffix":         users.Suffix,
			"email":          users.Email,
			"program":        users.Program,
			"year_level":     users.YearLevel,
			"contact_number": users.ContactNumber,
			"avatar_path": func() interface{} {
				if users.AvatarPath != "" {
					return users.AvatarPath
				}
				return nil
			}(),
			"redirect_url":    redirectURL,
			"borrowed_books":  borrowedBooksWithDetails,
		},
	})
}

func LogOutUser(c *fiber.Ctx) error {
	// Invalidate the cookie by setting MaxAge to -1
	c.Cookie(&fiber.Cookie{
		Name:     "user_session",
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false, // match your login config (true if HTTPS)
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Logout successful",
	})
}


// EditUser updates a user's personal information
func EditUser(c *fiber.Ctx) error {
	var request struct {
		UserID     string `json:"user_id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MiddleName string `json:"middle_name"`
		Suffix     string `json:"suffix"`
		ContactNumber string `json:"contact_number"`
		Program       string `json:"program"`
		YearLevel     string `json:"year_level"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "Invalid request payload",
			"Error":   err.Error(),
		})
	}

	// Find user by ID
	var user model.User
	if err := middleware.DBConn.Table("users").Where("user_id = ?", request.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"RetCode": "404",
			"Message": "User not found",
		})
	}

	// Update user fields
	user.FirstName = request.FirstName
	user.LastName = request.LastName
	user.MiddleName = &request.MiddleName
	user.Suffix = &request.Suffix
	user.ContactNumber = &request.ContactNumber
	user.Program = &request.Program
	user.YearLevel = &request.YearLevel

	// Save changes
	if err := middleware.DBConn.Table("users").Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to update user",
			"Error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"RetCode": "200",
		"Message": "User updated successfully",
		"Data":    user,
	})
}

func UpdateAvatar(c *fiber.Ctx) error {
	type AvatarUpdateRequest struct {
		UserID     string `json:"user_id"`
		AvatarPath string `json:"avatar_path"`
	}

	var req AvatarUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "Invalid JSON",
		})
	}

	if err := middleware.DBConn.Table("users").
		Where("user_id = ?", req.UserID).
		Update("avatar_path", req.AvatarPath).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to update avatar",
			"Error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"RetCode": "200",
		"Message": "Avatar updated successfully",
	})
}





func ReserveBook(c *fiber.Ctx) error {
    var reservation model.Reservation

    // Ensure the request body is correctly parsed
    if err := c.BodyParser(&reservation); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
            RetCode: "400",
            Message: "Invalid request body",
            Data:    err.Error(),
        })
    }

    // Check if the book exists
    var book model.Book
    if err := middleware.DBConn.First(&book, reservation.BookID).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
            RetCode: "404",
            Message: "Book not found",
            Data:    nil,
        })
    }

    // Check if the user already has an active reservation for the same book
    var existingReservation model.Reservation
    if err := middleware.DBConn.Where("user_id = ? AND book_id = ? AND (status = 'Pending' OR status = 'Approved')", reservation.UserID, reservation.BookID).First(&existingReservation).Error; err == nil {
        return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
            RetCode: "409",
            Message: "User already has an active reservation for this book",
            Data:    nil,
        })
    }

	// Clean up expired reservations
	middleware.DBConn.
	Where("status = ? AND expiry < ?", "Pending", time.Now()).
	Delete(&model.Reservation{})
    // Generate a unique ReservationID
    reservation.ReservationID = int(time.Now().UnixNano() % 10000)
    reservation.Status = "Pending"
    reservation.CreatedAt = time.Now()

	reservation.Expiry = reservation.PreferredPickupDate.Add(24 * time.Hour)

	// Optional: Set to expired if already expired on creation
	if time.Now().After(reservation.Expiry) {
		reservation.Status = "Expired"
	}

    // Create the reservation
    if err := middleware.DBConn.Create(&reservation).Error; err != nil {
        fmt.Println("DB Create Error:", err) // Log the error for debugging
        return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
            RetCode: "500",
            Message: "Failed to create reservation",
            Data:    err.Error(),
        })
    }

    return c.JSON(response.ResponseModel{
        RetCode: "201",
        Message: "Reservation successful. Pending approval.",
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

func FetchBorrowedBooks(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "User ID is required",
			Data:    nil,
		})
	}

	// Updated BorrowedBookData struct to include additional fields
	type BorrowedBookData struct {
		BookID       int       `json:"book_id"`
		Title        string    `json:"title"`
		Picture      string    `json:"picture"`
		Copies       int       `json:"copies"`
		DueDate      time.Time `json:"due_date"`
		Author       string    `json:"author"`       // New field
		Year         string    `json:"year_published"`         // New field
		ISBN         string    `json:"isbn"`         // New field
		ShelfLocation string    `json:"shelf_location"` // New field
		LibrarySection string    `json:"library_section"` // New field
		Description  string    `json:"description"`    // New field
	}

	var books []BorrowedBookData
	err := middleware.DBConn.Table("borrowed_books").
		Select("books.book_id, books.title, books.picture, books.available_copies AS copies, borrowed_books.due_date, books.author, books.year_published, books.isbn, books.shelf_location, books.library_section, books.description").
		Joins("JOIN books ON books.book_id = borrowed_books.book_id").
		Where("borrowed_books.user_id = ? AND borrowed_books.status = ?", userID, "Approved").
		Scan(&books).Error

	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch borrowed books",
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
	type BookWithReservedCount struct {
		model.Book
		ReservedCount int `json:"reserved_count"`
	}

	var booksWithCount []BookWithReservedCount

	if err := middleware.DBConn.
		Table("books").
		Select("books.*, COUNT(reservations.reservation_id) as reserved_count").
		Joins("LEFT JOIN reservations ON reservations.book_id = books.book_id AND reservations.status = ?", "Pending").
		Group("books.book_id").
		Scan(&booksWithCount).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to Fetch All Books",
			Data:    err.Error(),
		})
	}

	// Add base64 prefix to picture
	for i := range booksWithCount {
		if booksWithCount[i].Picture != "" {
			booksWithCount[i].Picture = "data:image/jpeg;base64," + booksWithCount[i].Picture
		}
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "All Books Fetched Successfully",
		Data:    booksWithCount,
	})
}






