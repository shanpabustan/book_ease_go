package controller

import (
	"book_ease_go/middleware"
	"book_ease_go/model"
	"book_ease_go/notifications"
	response "book_ease_go/responses"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	for _, user := range students {
		go func(u model.User) {
			notifications.NotifyAdminNewUser(middleware.DBConn, u)
		}(user)
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
		"user_id":   userID,
		"user_type": userType,
		"exp":       time.Now().Add(24 * time.Hour).Unix(), // Token expiration time (24 hours)
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
		// Check if this is a student account
		if users.UserType == "Student" {
			// Check if semester end date is set
			var setting model.Setting
			if err := middleware.DBConn.Where("key = ?", "semester_end_date").First(&setting).Error; err == nil {
				// If semester end date exists, return semester end message
				return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
					RetCode: "403",
					Message: "Your account is currently disabled due to semester end.",
				})
			}
		}
		// For other cases (like penalty), return the original message
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "The account is blocked due to penalty.",
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
				books.year_published,
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
			"redirect_url":   redirectURL,
			"borrowed_books": borrowedBooksWithDetails,
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
		UserID        string `json:"user_id"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		Email         string `json:"email"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
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
	user.UserID = request.UserID
	user.FirstName = request.FirstName
	user.LastName = request.LastName
	user.Email = request.Email
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
		Where("status = ? AND expired_at < ?", "Pending", time.Now()).
		Delete(&model.Reservation{})

	reservation.ReservationID = int(time.Now().Unix()%1000000000) + rand.Intn(10000)
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

	var user model.User
	if err := middleware.DBConn.Table("users").Where("user_id = ?", reservation.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user details",
			Data:    err.Error(),
		})
	}

	// Send notifications
	go func() {
		// Notify user about pending reservation
		notifications.NotifyPendingReservation(middleware.DBConn, user, book)
		// Notify admins about new reservation request
		notifications.NotifyAdminReservationRequest(middleware.DBConn, user, book)
	}()

	return c.JSON(response.ResponseModel{
		RetCode: "201",
		Message: "Reservation successful. Pending approval.",
		Data:    reservation,
	})
}

//FETCHING SECTION

// CourseToCategories maps courses to relevant book categories
var CourseToCategories = map[string][]string{
	"BS Computer Science":        {"Computer Science", "Information System", "Science & Technology", "Textbooks"},
	"BS Information Technology":  {"Information System", "Computer Science", "Science & Technology", "Textbooks"},
	"BS Business Administration": {"Business Administration", "Textbooks", "Reference Materials"},
	"BS Engineering":             {"Engineering", "Science & Technology", "Textbooks"},
	"BS Education":               {"Education", "Textbooks", "Reference Materials"},
	"BS Accountancy":             {"Accountancy", "Business Administration", "Textbooks"},
	"BS Psychology":              {"Psychology", "Non-Fiction", "Textbooks"},
	"BS Nursing":                 {"Nursing", "Biology", "Science & Technology", "Textbooks"},
	"BS Criminology":             {"Criminology", "History & Social Studies", "Textbooks"},
	"BS Hospitality Management":  {"Hospitality Management", "Business Administration", "Textbooks"},
	"BS Tourism Management":      {"Tourism Management", "Business Administration", "Textbooks"},
	"BS Architecture":            {"Architecture", "Engineering", "Textbooks"},
	"BS Civil Engineering":       {"Civil Engineering", "Engineering", "Textbooks"},
	"BS Mechanical Engineering":  {"Mechanical Engineering", "Engineering", "Textbooks"},
	"BS Electrical Engineering":  {"Electrical Engineering", "Engineering", "Textbooks"},
	"BS Electronics Engineering": {"Electronics Engineering", "Engineering", "Textbooks"},
	"BS Pharmacy":                {"Pharmacy", "Biology", "Science & Technology", "Textbooks"},
	"BS Biology":                 {"Biology", "Science & Technology", "Textbooks"},
	"BS Mathematics":             {"Mathematics", "Textbooks"},
	"BS Environmental Science":   {"Environmental Science", "Biology", "Science & Technology", "Textbooks"},
	"AB Communication":           {"Communication", "Non-Fiction", "Textbooks"},
	"AB Political Science":       {"Political Science", "History & Social Studies", "Textbooks"},
	"AB English":                 {"English", "Fiction", "Non-Fiction", "Textbooks"},
	"AB History":                 {"History", "History & Social Studies", "Biographies", "Textbooks"},
}

// FetchRecommendedBooks retrieves books based on the user's course
func FetchRecommendedBooks(c *fiber.Ctx) error {
	// Get user_id from query parameter
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "User ID is required",
			"Error":   "user_id query parameter is missing",
		})
	}

	// Fetch user to get their Program
	var user model.User
	if err := middleware.DBConn.Where("user_id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"RetCode": "404",
				"Message": "User not found or inactive",
				"Error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to fetch user",
			"Error":   err.Error(),
		})
	}

	// Check if user has a Program
	if user.Program == nil || *user.Program == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "User has no associated program",
			"Error":   "Program field is empty",
		})
	}

	// Get categories for the user's program
	categories, exists := CourseToCategories[*user.Program]
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"RetCode": "400",
			"Message": "No recommendations available for this program",
			"Error":   "Program not found in category mapping",
		})
	}

	// Fetch books matching the categories
	var books []model.Book
	if err := middleware.DBConn.Where("category IN ?", categories).
		Where("available_copies > ?", 0).
		// Select("book_id", "title", "author", "category", "isbn", "available_copies", "picture").
		Find(&books).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"RetCode": "500",
			"Message": "Failed to fetch recommended books",
			"Error":   err.Error(),
		})
	}

	// Return empty list if no books found
	if len(books) == 0 {
		return c.JSON(fiber.Map{
			"RetCode": "200",
			"Message": "No recommended books found for your program",
			"Data":    []model.Book{},
		})
	}

	// Return recommended books
	return c.JSON(fiber.Map{
		"RetCode": "200",
		"Message": "Recommended books retrieved successfully",
		"Data":    books,
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

	type BorrowedBookData struct {
		BookID         int       `json:"book_id"`
		Title          string    `json:"title"`
		Picture        string    `json:"picture"`
		Copies         int       `json:"copies"`
		DueDate        time.Time `json:"due_date"`
		Author         string    `json:"author"`
		YearPublished  string    `json:"year_published"` // Updated field name to match database
		ISBN           string    `json:"isbn"`
		ShelfLocation  string    `json:"shelf_location"`
		LibrarySection string    `json:"library_section"`
		Description    string    `json:"description"`
	}

	var books []BorrowedBookData
	err := middleware.DBConn.Debug(). // Added Debug() to see the SQL query
						Table("borrowed_books").
						Select(`
			books.book_id, 
			books.title, 
			books.picture, 
			books.available_copies AS copies, 
			borrowed_books.due_date, 
			books.author, 
			books.year_published AS year_published, 
			books.isbn, 
			books.shelf_location, 
			books.library_section, 
			books.description
		`).
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

func FetchBorrowedBooksByStatus(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	status := c.Query("status")

	if userID == "" {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "User ID is required",
			Data:    nil,
		})
	}

	type BookData struct {
		BookID        int       `json:"book_id"`
		Title         string    `json:"title"`
		Picture       string    `json:"picture"`
		BorrowDate    time.Time `json:"borrow_date"`
		DueDate       time.Time `json:"due_date"`
		Status        string    `json:"status"`
		Source        string    `json:"source"`
		ReservationID int       `json:"reservation_id,omitempty"`
	}

	var books []BookData

	var query *gorm.DB
	if status == "" || status == "All" {
		// Query for all statuses
		query = middleware.DBConn.Raw(`
			(SELECT 
				books.book_id,
				books.title,
				books.picture,
				borrowed_books.borrow_date,
				borrowed_books.due_date,
				CASE 
					WHEN borrowed_books.status = 'Pending' THEN 'To Return'
					WHEN borrowed_books.status = 'Cancelled' THEN 'Cancelled'
					WHEN borrowed_books.status = 'Returned' THEN 'Returned'
					ELSE borrowed_books.status 
				END as status,
				'borrowed' as source,
				NULL as reservation_id
			FROM borrowed_books
			JOIN books ON books.book_id = borrowed_books.book_id
			WHERE borrowed_books.user_id = ?)
			
			UNION ALL
			
			(SELECT 
				books.book_id,
				books.title,
				books.picture,
				reservations.created_at as borrow_date,
    			reservations.expired_at as due_date,  
				CASE 
					WHEN reservations.status = 'Pending' THEN 'To Pick Up'
					WHEN reservations.status = 'Cancelled' THEN 'Cancelled'
					WHEN reservations.status = 'Picked Up' THEN 'Picked Up'
					ELSE reservations.status 
				END as status,
				'reservation' as source,
				reservations.reservation_id
			FROM reservations
			JOIN books ON books.book_id = reservations.book_id
			WHERE reservations.user_id = ? AND reservations.status != 'Approved')
			
			ORDER BY borrow_date DESC
		`, userID, userID)
	} else {
		// Handle specific status mapping
		var borrowedStatus, reservationStatus string

		switch status {
		case "To Return":
			borrowedStatus = "Pending"
			reservationStatus = "none" // Won't match any reservation
		case "To Pick Up":
			borrowedStatus = "none" // Won't match any borrowed book
			reservationStatus = "Pending"
		default:
			borrowedStatus = status
			reservationStatus = status
		}

		query = middleware.DBConn.Raw(`
			(SELECT 
				books.book_id,
				books.title,
				books.picture,
				borrowed_books.borrow_date,
				borrowed_books.due_date,
				CASE 
					WHEN borrowed_books.status = 'Pending' THEN 'To Return'
					WHEN borrowed_books.status = 'Cancelled' THEN 'Cancelled'
					WHEN borrowed_books.status = 'Returned' THEN 'Returned'
					ELSE borrowed_books.status 
				END as status,
				'borrowed' as source,
				NULL as reservation_id
			FROM borrowed_books
			JOIN books ON books.book_id = borrowed_books.book_id
			WHERE borrowed_books.user_id = ? AND borrowed_books.status = ?)
			
			UNION ALL
			
			(SELECT 
				books.book_id,
				books.title,
				books.picture,
				reservations.created_at as borrow_date,
				reservations.expired_at as due_date,
				CASE 
					WHEN reservations.status = 'Pending' THEN 'To Pick Up'
					WHEN reservations.status = 'Cancelled' THEN 'Cancelled'
					WHEN reservations.status = 'Picked Up' THEN 'Picked Up'
					ELSE reservations.status 
				END as status,
				'reservation' as source,
				reservations.reservation_id
			FROM reservations
			JOIN books ON books.book_id = reservations.book_id
			WHERE reservations.user_id = ? AND reservations.status = ? AND reservations.status != 'Approved')
			
			ORDER BY borrow_date DESC
		`, userID, borrowedStatus, userID, reservationStatus)
	}

	if err := query.Scan(&books).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch books and reservations",
			Data:    err.Error(),
		})
	}

	// Add base64 prefix to picture if it exists
	for i := range books {
		if books[i].Picture != "" {
			books[i].Picture = "data:image/jpeg;base64," + books[i].Picture
		}
	}

	message := "Books and Reservations Fetched Successfully"
	if status != "" && status != "all" {
		message = fmt.Sprintf("Books and Reservations with status '%s' Fetched Successfully", status)
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: message,
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

func ChangePassword(c *fiber.Ctx) error {
	// Define request structure
	var request struct {
		UserID          string `json:"user_id"`
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    err.Error(),
		})
	}

	// Validate request fields
	if request.UserID == "" || request.CurrentPassword == "" || request.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "All fields are required",
		})
	}

	// Find user in database
	var user model.User
	if err := middleware.DBConn.Table("users").Where("user_id = ?", request.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
		})
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.CurrentPassword)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Current password is incorrect",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash new password",
			Data:    err.Error(),
		})
	}

	// Update password in database
	if err := middleware.DBConn.Table("users").Where("user_id = ?", request.UserID).Update("password", string(hashedPassword)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update password",
			Data:    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Password updated successfully",
	})
}

func FetchBooksAllStatus(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	if userID == "" {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "User ID is required",
			Data:    nil,
		})
	}

	type BookData struct {
		BookID     int       `json:"book_id"`
		Title      string    `json:"title"`
		Picture    string    `json:"picture"`
		BorrowDate time.Time `json:"borrow_date"`
		DueDate    time.Time `json:"due_date"`
		Status     string    `json:"status"`
	}

	var books []BookData

	// Create the union query without status filter
	query := middleware.DBConn.Raw(`
		(SELECT 
			books.book_id,
			books.title,
			books.picture,
			borrowed_books.borrow_date,
			borrowed_books.due_date,
			borrowed_books.status
		FROM borrowed_books
		JOIN books ON books.book_id = borrowed_books.book_id
		WHERE borrowed_books.user_id = ?)
		
		UNION ALL
		
		(SELECT 
			books.book_id,
			books.title,
			books.picture,
			reservations.created_at as borrow_date,
			reservations.preferred_pickup_date as due_date,
			reservations.status
		FROM reservations
		JOIN books ON books.book_id = reservations.book_id
		WHERE reservations.user_id = ?)
		
		ORDER BY borrow_date DESC
	`, userID, userID)

	if err := query.Scan(&books).Error; err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch books and reservations",
			Data:    err.Error(),
		})
	}

	// Add base64 prefix to picture if it exists
	for i := range books {
		if books[i].Picture != "" {
			books[i].Picture = "data:image/jpeg;base64," + books[i].Picture
		}
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "All Books and Reservations Fetched Successfully",
		Data:    books,
	})
}

// FetchMostPopularBooks returns the most popular books based on borrow count
func FetchMostPopularBooks(c *fiber.Ctx) error {
	// Get limit from query parameter, default to 10 if not provided
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	type PopularBook struct {
		BookID          int       `json:"book_id"`
		Title           string    `json:"title"`
		Author          string    `json:"author"`
		Category        string    `json:"category"`
		ISBN            string    `json:"isbn"`
		LibrarySection  string    `json:"library_section"`
		ShelfLocation   string    `json:"shelf_location"`
		TotalCopies     int       `json:"total_copies"`
		AvailableCopies int       `json:"available_copies"`
		BookCondition   string    `json:"book_condition"`
		Picture         string    `json:"picture"`
		YearPublished   int       `json:"year_published"`
		Version         int       `json:"version"`
		Description     string    `json:"description"`
		BorrowCount     int       `json:"borrow_count"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var popularBooks []PopularBook

	// Query to get most borrowed books with their details
	query := `
		SELECT 
			b.book_id,
			b.title,
			b.author,
			b.category,
			b.isbn,
			b.library_section,
			b.shelf_location,
			b.total_copies,
			b.available_copies,
			b.book_condition,
			b.picture,
			b.year_published,
			b.version,
			b.description,
			b.created_at,
			COUNT(bb.borrow_id) as borrow_count
		FROM books b
		LEFT JOIN borrowed_books bb ON b.book_id = bb.book_id
		GROUP BY b.book_id, b.title, b.author, b.category, b.isbn, 
			b.library_section, b.shelf_location, b.total_copies, 
			b.available_copies, b.book_condition, b.picture, 
			b.year_published, b.version, b.description, b.created_at
		ORDER BY borrow_count DESC
		LIMIT ?
	`

	if err := middleware.DBConn.Raw(query, limit).Scan(&popularBooks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch popular books",
			Data:    err.Error(),
		})
	}

	// Add base64 prefix to picture if it exists
	for i := range popularBooks {
		if popularBooks[i].Picture != "" {
			popularBooks[i].Picture = "data:image/jpeg;base64," + popularBooks[i].Picture
		}
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Most popular books fetched successfully",
		Data:    popularBooks,
	})
}

// RequestPasswordReset handles the initial password reset request
func RequestPasswordReset(c *fiber.Ctx) error {
	var request struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    err.Error(),
		})
	}

	// Validate email
	if request.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Email is required",
		})
	}

	// Check if user exists
	var user model.User
	if err := middleware.DBConn.Table("users").Where("email = ?", request.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "If your email is registered, you will receive a reset code",
		})
	}

	// Generate a 6-digit code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiresAt := time.Now().Add(3 * time.Minute)

	// Save the reset code
	resetCode := model.PasswordResetCode{
		Email:     request.Email,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	if err := middleware.DBConn.Create(&resetCode).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to generate reset code",
			Data:    err.Error(),
		})
	}

	// Send email with reset code
	subject := "Password Reset Code - Book Ease"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #008080;">Password Reset Request</h2>
			<p>You have requested to reset your password. Use the following code to proceed:</p>
			<div style="background-color: #f5f5f5; padding: 15px; border-radius: 5px; text-align: center; font-size: 24px; letter-spacing: 5px; margin: 20px 0;">
				<strong>%s</strong>
			</div>
			<p>This code will expire in 3 minutes.</p>
			<p>If you did not request this reset, please ignore this email.</p>
			<hr style="border: 1px solid #008080;">
			<p style="color: #666; font-size: 12px;">This is an automated message, please do not reply.</p>
		</div>
	`, code)

	if err := notifications.SendEmail(request.Email, subject, body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to send reset code",
			Data:    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "If your email is registered, you will receive a reset code",
	})
}

// VerifyResetCodeAndResetPassword handles the code verification and password reset
func VerifyResetCodeAndResetPassword(c *fiber.Ctx) error {
	var request struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    err.Error(),
		})
	}

	// Validate request fields
	if request.Email == "" || request.Code == "" || request.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Email, code, and new password are required",
		})
	}

	// Find the reset code
	var resetCode model.PasswordResetCode
	if err := middleware.DBConn.Where("email = ? AND code = ? AND used = ? AND expires_at > ?",
		request.Email, request.Code, false, time.Now()).First(&resetCode).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid or expired reset code",
		})
	}

	// Find user
	var user model.User
	if err := middleware.DBConn.Table("users").Where("email = ?", request.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to hash new password",
			Data:    err.Error(),
		})
	}

	// Update password in database
	if err := middleware.DBConn.Table("users").Where("email = ?", request.Email).Update("password", string(hashedPassword)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update password",
			Data:    err.Error(),
		})
	}

	// Mark reset code as used
	middleware.DBConn.Model(&resetCode).Update("used", true)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Password reset successful",
	})
}
