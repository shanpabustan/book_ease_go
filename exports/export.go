package export

import (
	"book_ease_go/model"
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func generateTimestampFilename(prefix string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.pdf", prefix, timestamp)
}

func derefString(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}

func ExportBooksCSV(books []model.Book) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write([]string{"BookID", "Title", "Author", "Category", "ISBN", "LibrarySection", "ShelfLocation", "TotalCopies", "AvailableCopies", "BookCondition", "YearPublished", "Version", "Description"})

	for _, book := range books {
		record := []string{
			fmt.Sprintf("%d", book.BookID),
			book.Title,
			book.Author,
			book.Category,
			book.ISBN,
			book.LibrarySection,
			book.ShelfLocation,
			fmt.Sprintf("%d", book.TotalCopies),
			fmt.Sprintf("%d", book.AvailableCopies),
			book.BookCondition,
			fmt.Sprintf("%d", book.YearPublished),
			fmt.Sprintf("%d", book.Version),
			book.Description,
		}
		writer.Write(record)
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportBooksExcel(books []model.Book) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := []string{"BookID", "Title", "Author", "Category", "ISBN", "LibrarySection", "ShelfLocation", "TotalCopies", "AvailableCopies", "BookCondition", "YearPublished", "Version", "Description"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheet, cell, h)
	}

	for i, book := range books {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), book.BookID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), book.Title)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2), book.Author)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i+2), book.Category)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+2), book.ISBN)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+2), book.LibrarySection)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+2), book.ShelfLocation)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", i+2), book.TotalCopies)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", i+2), book.AvailableCopies)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", i+2), book.BookCondition)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", i+2), book.YearPublished)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", i+2), book.Version)
		f.SetCellValue(sheet, fmt.Sprintf("M%d", i+2), book.Description)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportBooksPDF(books []model.Book) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)

	pdf.Cell(20, 10, "ID")
	pdf.Cell(60, 10, "Title")
	pdf.Cell(40, 10, "Author")
	pdf.Cell(40, 10, "Category")
	pdf.Ln(10)

	for _, book := range books {
		pdf.Cell(20, 10, fmt.Sprintf("%d", book.BookID))
		pdf.Cell(60, 10, book.Title)
		pdf.Cell(40, 10, book.Author)
		pdf.Cell(40, 10, book.Category)
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportUsersCSV(users []model.User) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write([]string{"UserID", "FirstName", "LastName", "Email", "Program", "YearLevel", "IsActive"})

	for _, user := range users {
		record := []string{
			user.UserID,
			user.FirstName,
			user.LastName,
			user.Email,
			derefString(user.Program),
			derefString(user.YearLevel),
			fmt.Sprintf("%t", user.IsActive),
		}
		writer.Write(record)
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportUsersExcel(users []model.User) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := []string{"UserID", "FirstName", "LastName", "Email", "Program", "YearLevel", "IsActive"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheet, cell, h)
	}

	for i, user := range users {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), user.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), user.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2), user.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i+2), user.Email)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+2), derefString(user.Program))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+2), derefString(user.YearLevel))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+2), user.IsActive)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportUsersPDF(users []model.User) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 12)

	pdf.Cell(30, 10, "UserID")
	pdf.Cell(40, 10, "FirstName")
	pdf.Cell(40, 10, "LastName")
	pdf.Cell(60, 10, "Email")
	pdf.Cell(40, 10, "Program")
	pdf.Cell(30, 10, "YearLevel")
	pdf.Ln(10)

	for _, user := range users {
		pdf.Cell(30, 10, user.UserID)
		pdf.Cell(40, 10, user.FirstName)
		pdf.Cell(40, 10, user.LastName)
		pdf.Cell(60, 10, user.Email)
		pdf.Cell(40, 10, derefString(user.Program))
		pdf.Cell(30, 10, derefString(user.YearLevel))
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
