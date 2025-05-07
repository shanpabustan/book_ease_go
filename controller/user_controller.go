package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	"book_ease_go/notifications"
	response "book_ease_go/responses"
	"bytes"
	"errors"
	"fmt"
	"os"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register User
func CreateStudent(c *fiber.Ctx) error {
	var students []model.User

	// First try: parse as array
	if err := c.BodyParser(&students); err != nil {
		// Second try: parse as single user
		var single model.User	
		if err := c.BodyParser(&single); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RetCode": "400",
				"Message": "Invalid request data",
				"Error":   err.Error(),
			})
		}
		students = append(students, single)
	}

	for i := range students {
		// Force user_type and default avatar path
		students[i].UserType = "Student"
		students[i].AvatarPath = ""

		// Check for duplicate user_id
		var existing model.User
		err := middleware.DBConn.Table("users").Where("user_id = ?", students[i].UserID).First(&existing).Error
		if err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RetCode": "400",
				"Message": fmt.Sprintf("User ID already exists: %s", students[i].UserID),
			})
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"RetCode": "500",
				"Message": "Database error",
				"Error":   err.Error(),
			})
		}

		// Hash the password
		hashed, err := bcrypt.GenerateFromPassword([]byte(students[i].Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"RetCode": "500",
				"Message": "Failed to hash password",
				"Error":   err.Error(),
			})
		}
		students[i].Password = string(hashed)
	}

	// Insert users
	if err := middleware.DBConn.Table("users").Create(&students).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to create user(s)",
			"Error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"RetCode": "201",
		"Message": "Registration Successful",
		"Data":    students,
	})
}


// Secret key for signing JWT tokens (consider storing it securely, like in an environment variable)
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID string, userType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"user_type": userType,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token expiration time (24 hours)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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

	if !users.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "The account is blocked. Please contact the administrator.",
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

	if borrowedBooksWithDetails == nil {
		borrowedBooksWithDetails = []model.BorrowedBookWithDetails{}
	}

	// ✅ Generate JWT
	token, err := GenerateJWT(users.UserID, users.UserType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Error generating token",
		})
	}

	// ✅ Set JWT in cookie
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // Change to true if using HTTPS
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
	// Invalidate the "jwt" cookie to log the user out
	c.Cookie(&fiber.Cookie{
		Name:     "jwt", // ✅ Match the name set in login
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false,
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

	if err := c.BodyParser(&reservation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    err.Error(),
		})
	}

	var book model.Book
	if err := middleware.DBConn.First(&book, reservation.BookID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Book not found",
			Data:    nil,
		})
	}
	
	if book.AvailableCopies <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Cannot reserve. No available copies at the moment.",
			Data:    nil,
		})
	}

	var existingReservation model.Reservation
	if err := middleware.DBConn.Where("user_id = ? AND book_id = ? AND (status = 'Pending')", reservation.UserID, reservation.BookID).First(&existingReservation).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
			RetCode: "409",
			Message: "User already has an active reservation for this book",
			Data:    nil,
		})
	}


	middleware.DBConn.
		Where("status = ? AND expiry < ?", "Pending", time.Now()).
		Delete(&model.Reservation{})

	reservation.ReservationID = int(time.Now().UnixNano() % 10000)
	reservation.Status = "Pending"
	reservation.CreatedAt = time.Now()
	reservation.Expiry = reservation.PreferredPickupDate.Add(24 * time.Hour)

	if time.Now().After(reservation.Expiry) {
		reservation.Status = "Expired"
	}

	if err := middleware.DBConn.Create(&reservation).Error; err != nil {
		fmt.Println("DB Create Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to create reservation",
			Data:    err.Error(),
		})
	}

	// Generate notification message using template
	var msgBuffer bytes.Buffer
	err := notifications.NotificationTemplates["ReservationPending"].Execute(&msgBuffer, map[string]string{
		"BookTitle": book.Title,
	})
	if err != nil {
		fmt.Println("Template execution error:", err)
	}

	notification := model.Notification{
		UserID:  reservation.UserID,
		Message: msgBuffer.String(),
		IsRead:  false,
	}

	if err := middleware.DBConn.Create(&notification).Error; err != nil {
		fmt.Println("Error sending notification:", err)
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
		Where("borrowed_books.user_id = ? AND borrowed_books.status = ?", userID, "Pending").
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






