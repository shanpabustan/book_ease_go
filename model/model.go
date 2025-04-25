package model

import (
	"time"

	
)
	type User struct {
		UserID        string    `gorm:"primaryKey;size:20" json:"user_id"`
		UserType      string    `gorm:"size:20;check:user_type IN ('Admin','Student')" json:"user_type"`
		LastName      string    `gorm:"size:255;not null" json:"last_name"`
		FirstName     string    `gorm:"size:255;not null" json:"first_name"`
		MiddleName    *string   `gorm:"size:255" json:"middle_name,omitempty"`
		Suffix        *string   `gorm:"size:4" json:"suffix,omitempty"`
		Email         string    `gorm:"size:255;unique;not null" json:"email"`
		Password  string    `gorm:"type:text;not null" json:"password"`
		Program       *string   `gorm:"size:50" json:"program,omitempty"`
		YearLevel     *string   `gorm:"size:20" json:"year_level,omitempty"`
		ContactNumber *string   `gorm:"size:20" json:"contact_number,omitempty"`
		AvatarPath    string    `gorm:"size:255" json:"avatar_path"`
		IsActive bool `gorm:"default:true" json:"is_active"`
		CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

		// Relationships
		Reservations   []Reservation   `gorm:"foreignKey:UserID"`
		BorrowedBooks  []BorrowedBook  `gorm:"foreignKey:UserID"`
		Notifications  []Notification  `gorm:"foreignKey:UserID"`
	}



type Book struct {
	BookID          int       `gorm:"primaryKey" json:"book_id"`
	Title           string    `gorm:"not null" json:"title"`
	Author          string    `gorm:"not null" json:"author"`
	Category        string    `gorm:"not null" json:"category"`
	ISBN            string    `gorm:"not null" json:"isbn"`
	LibrarySection  string    `gorm:"not null" json:"library_section"`
	ShelfLocation   string    `gorm:"not null" json:"shelf_location"`
	TotalCopies     int       `gorm:"not null" json:"total_copies"`
	AvailableCopies int       `gorm:"not null" json:"available_copies"`
	BookCondition   string    `gorm:"not null" json:"book_condition"`
	Picture         string    `gorm:"type:text" json:"picture"` // base64-encoded image
	YearPublished   int       `gorm:"not null" json:"year_published"`
	Version         int       `gorm:"not null" json:"version"`
	Description     string    `gorm:"type:text" json:"description"`

	// Relationships
	Reservations   []Reservation   `gorm:"foreignKey:BookID" json:"reservations,omitempty"`
	BorrowedBooks  []BorrowedBook  `gorm:"foreignKey:BookID" json:"borrowed_books,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Reservation struct {
	ReservationID       int       `gorm:"primaryKey" json:"reservation_id"`
	UserID              string    `gorm:"size:20;not null" json:"user_id"` 
	BookID              int       `gorm:"not null" json:"book_id"`
	PreferredPickupDate time.Time `gorm:"not null" json:"preferred_pickup_date"`
	Expiry              time.Time `gorm:"column:expired_at;not null" json:"expired_at"`
	Status 				string 	`gorm:"type:varchar(50);default:'Pending';check:status IN ('Pending', 'Approved', 'Cancelled', 'Expired')" json:"status"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	Book Book `gorm:"foreignKey:BookID;references:BookID;constraint:OnDelete:CASCADE"`
}
type BorrowedBook struct {
    BorrowID            int        `gorm:"primaryKey;column:borrow_id" json:"borrow_id"`
    ReservationID       int        `gorm:"not null" json:"reservation_id"`
    UserID              string     `gorm:"size:20;not null" json:"user_id"`
    BookID              int        `gorm:"not null" json:"book_id"`
    BorrowDate          time.Time  `gorm:"autoCreateTime" json:"borrow_date"`
    DueDate             time.Time  `gorm:"not null" json:"due_date"`
    ReturnDate          *time.Time `json:"return_date,omitempty"`
    Status              string     `gorm:"type:varchar(50);default:'Pending';check:status IN ('Pending', 'Approved', 'Returned', 'Overdue', 'Damaged')" json:"status"`
    BookConditionBefore string     `gorm:"type:varchar(20);check:book_condition_before IN ('New', 'Good', 'Fair', 'Poor')" json:"book_condition_before"`
    BookConditionAfter  *string    `gorm:"type:varchar(20);check:book_condition_after IN ('New', 'Good', 'Fair', 'Poor', 'Damaged')" json:"book_condition_after"`
    PenaltyAmount       float64    `gorm:"type:decimal(10,2);default:0" json:"penalty_amount"`
    
    // Relationships
    User                User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
    Book                Book        `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE" json:"book"`
    Reservation         Reservation `gorm:"foreignKey:ReservationID;constraint:OnDelete:CASCADE" json:"reservation"`
}

type BorrowedBookWithDetails struct {
	ReservationID int      `json:"reservation_id"`
	UserID        string    `json:"user_id"`
	BookID        int      `json:"book_id"`
	Title         string    `json:"title"`
	Picture       string    `json:"picture"`      // <--- This matches the alias used in the SELECT
	BorrowDate    time.Time `json:"borrow_date"`
	DueDate       time.Time `json:"due_date"`
}


type Notification struct {
	NotificationID uint      `gorm:"primaryKey;autoIncrement"`
	UserID         string    `gorm:"size:20;not null"`
	Message        string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	IsRead         bool      `gorm:"default:false"`
	User           User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Setting struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Key   string `gorm:"uniqueIndex;not null" json:"key"`
	Value string `gorm:"not null" json:"value"` // e.g., "2025-06-15"
}
