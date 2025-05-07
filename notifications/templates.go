// notifications/templates.go
package notifications

import (
	"text/template"
	
)

var NotificationTemplates = map[string]*template.Template{
	"DueDate": template.Must(template.New("DueDate").Parse(
		`The book "{{.BookTitle}}" must be returned on or before {{.DueDate}}. Please return it on time to avoid penalties.`,
	)),
	"ReservationApproved": template.Must(template.New("ReservationApproved").Parse(
    `Your reservation for the book "{{.BookTitle}}" has been approved. Please pick it up by {{.PreferredPickupDate}}.`,
	)),
	"ReservationDeclined": template.Must(template.New("ReservationDeclined").Parse(
		`We're sorry! Your reservation for the book "{{.BookTitle}}" has been declined. Please contact the librarian.`,
	)),
	"ReservationPending": template.Must(template.New("ReservationPending").Parse(
		`Your reservation for the book "{{.BookTitle}}" is pending. Please wait for approval.`,
	)),
	"ReservationCancelled": template.Must(template.New("ReservationCancelled").Parse(
		`Your reservation for the book "{{.BookTitle}}" has been cancelled. If you have any questions, please contact us.`,
	)),
	"ReservationExpired": template.Must(template.New("ReservationExpired").Parse(
		`Your reservation for the book "{{.BookTitle}}" has expired. Please try again or choose another book.`,
	)),
	"BookBorrowed": template.Must(template.New("BookBorrowed").Parse(
		`You have successfully borrowed the book "{{.BookTitle}}". The due date for return is {{.DueDate}}.`,
	)),
	"OverdueBook": template.Must(template.New("OverdueBook").Parse(
		`Your borrowed book "{{.BookTitle}}" is overdue. Please return it as soon as possible to avoid penalties.`,
	)),
	"BookReturned": template.Must(template.New("BookReturned").Parse(
		`Thank you for returning the book "{{.BookTitle}}". We hope you enjoyed it.`,
	)),
	"AccountActivated": template.Must(template.New("AccountActivated").Parse(
		`Your account has been activated successfully. You can now borrow books and manage your reservations.`,
	)),
	"AccountDeactivated": template.Must(template.New("AccountDeactivated").Parse(
		`Your account has been deactivated. Please contact the administrator for assistance.`,
	)),
	"NewReservationRequest": template.Must(template.New("NewReservationRequest").Parse(
		`A new reservation has been made by {{.UserName}} for the book "{{.BookTitle}}". Please review and approve or reject the reservation.`,
	)),
	"ReservationStatusChangedAdmin": template.Must(template.New("ReservationStatusChangedAdmin").Parse(
		`The reservation for "{{.BookTitle}}" by {{.UserName}} has been {{.Status}}. Please proceed with necessary actions.`,
	)),
}



