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
			book.Title,  // The Title can be long, so it will be fully exported.
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
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s1", col), h)
		f.SetCellStyle(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), style)
		// Increase the width for the "Title" column to 40, since it's expected to be longer
		if col == "B" {
			f.SetColWidth(sheet, col, col, 40)
		} else {
			f.SetColWidth(sheet, col, col, 18)
		}
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
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(79, 129, 189)
	pdf.SetTextColor(255, 255, 255)

	pdf.CellFormat(20, 10, "ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(90, 10, "Title", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Author", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Category", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)

	for _, book := range books {
		pdf.CellFormat(20, 10, fmt.Sprintf("%d", book.BookID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(90, 10, book.Title, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 10, book.Author, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 10, book.Category, "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
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

	headers := []string{"UserID", "FirstName", "LastName", "Email", "Program", "YearLevel"}
	writer.Write(headers)

	for _, user := range users {
		if user.UserType != "Student" {
			continue
		}
		record := []string{
			user.UserID,
			user.FirstName,
			user.LastName,
			user.Email,
			derefString(user.Program),
			derefString(user.YearLevel),
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
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"UserID", "FirstName", "LastName", "Email", "Program", "YearLevel"}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FF4F81BD"},
			Pattern: 1,
		},
	})

	for i, h := range headers {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, style)
		if h == "Program" {
			f.SetColWidth(sheet, col, col, 30)
		} else {
			f.SetColWidth(sheet, col, col, 18)
		}
	}

	row := 2
	for _, user := range users {
		if user.UserType != "Student" {
			continue
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), user.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), user.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), user.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), user.Email)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), derefString(user.Program))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), derefString(user.YearLevel))
		row++
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportUsersPDF(users []model.User) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(79, 129, 189)
	pdf.SetTextColor(255, 255, 255)

	headers := []struct {
		title string
		width float64
	}{
		{"UserID", 30},
		{"FirstName", 40},
		{"LastName", 40},
		{"Email", 60},
		{"Program", 60},
		{"YearLevel", 30},
	}

	for _, h := range headers {
		pdf.CellFormat(h.width, 10, h.title, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 11)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetTextColor(0, 0, 0)

	for _, user := range users {
		if user.UserType != "Student" {
			continue
		}
		pdf.CellFormat(30, 8, user.UserID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, user.FirstName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 8, user.LastName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 8, user.Email, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 8, derefString(user.Program), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, derefString(user.YearLevel), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
