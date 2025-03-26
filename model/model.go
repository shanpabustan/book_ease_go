package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UserID        string         `gorm:"primaryKey;size:20"`
	UserType      string         `gorm:"type:varchar(20);check:user_type IN ('Admin', 'Student')"`
	LastName      string         `gorm:"size:255;not null"`
	FirstName     string         `gorm:"size:255;not null"`
	MiddleName    *string        `gorm:"size:255"`
	Suffix        *string        `gorm:"size:4"`
	Email         string         `gorm:"size:255;unique;not null"`
	PasswordHash  string         `gorm:"not null"`
	Program       *string        `gorm:"size:50"`
	YearLevel     *string        `gorm:"size:20"`
	ContactNumber *string        `gorm:"size:20"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	Reservations  []Reservation  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	BorrowedBooks []BorrowedBook `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Notifications []Notification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

type Book struct {
	BookID          uint           `gorm:"primaryKey;autoIncrement"`
	Title           string         `gorm:"size:255;not null"`
	Author          string         `gorm:"size:255;not null"`
	Category        string         `gorm:"size:255;not null"`
	ISBN            string         `gorm:"size:20;unique;not null"`
	LibrarySection  string         `gorm:"size:255;not null"`
	ShelfLocation   string         `gorm:"size:50;not null"`
	TotalCopies     int            `gorm:"not null;check:total_copies >= 0"`
	AvailableCopies int            `gorm:"not null;check:available_copies >= 0"`
	BookCondition   string         `gorm:"type:varchar(10);check:book_condition IN ('New', 'Used')"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	Reservations    []Reservation  `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	BorrowedBooks   []BorrowedBook `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

type Reservation struct {
	ReservationID       uint      `gorm:"primaryKey;autoIncrement"`
	UserID              string    `gorm:"size:20;not null"`
	BookID              uint      `gorm:"not null"`
	PreferredPickupDate time.Time `gorm:"not null"`
	Status              string    `gorm:"type:varchar(50);default:'Pending';check:status IN ('Pending', 'Approved', 'Cancelled')"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	User                User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Book                Book      `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

type BorrowedBook struct {
	BorrowID            uint      `gorm:"primaryKey;autoIncrement"`
	UserID              string    `gorm:"size:20;not null"`
	BookID              uint      `gorm:"not null"`
	BorrowDate          time.Time `gorm:"autoCreateTime"`
	DueDate             time.Time `gorm:"not null"`
	ReturnDate          *time.Time
	Status              string  `gorm:"type:varchar(50);default:'Pending';check:status IN ('Pending', 'Approved', 'Returned', 'Overdue', 'Damaged')"`
	BookConditionBefore string  `gorm:"type:varchar(20);check:book_condition_before IN ('New', 'Good', 'Fair', 'Poor')"`
	BookConditionAfter  *string `gorm:"type:varchar(20);check:book_condition_after IN ('New', 'Good', 'Fair', 'Poor', 'Damaged')"`
	PenaltyAmount       float64 `gorm:"type:decimal(10,2);default:0"`
	User                User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Book                Book    `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

type Notification struct {
	NotificationID uint      `gorm:"primaryKey;autoIncrement"`
	UserID         string    `gorm:"size:20;not null"`
	Message        string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	IsRead         bool      `gorm:"default:false"`
	User           User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
