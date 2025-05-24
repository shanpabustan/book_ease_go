package main

import (
	"fmt"

	"book_ease_go/controller"
	"book_ease_go/middleware"
	"book_ease_go/model"
	"book_ease_go/notifications"
	"book_ease_go/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func init() {
	fmt.Println("STARTING SERVER...")
	fmt.Println("INITIALIZE DB CONNECTION...")
	if middleware.ConnectDB() {
		fmt.Println("DB CONNECTION FAILED!")
	} else {
		fmt.Println("DB CONNECTION SUCCESSFUL!")
	}
	// 🔄 Run auto-migration for User model
	middleware.DBConn.AutoMigrate(&model.User{})
	middleware.DBConn.AutoMigrate(&model.Book{})
	middleware.DBConn.AutoMigrate(&model.Reservation{})
	middleware.DBConn.AutoMigrate(&model.BorrowedBook{})
	middleware.DBConn.AutoMigrate(&model.Notification{})
	middleware.DBConn.AutoMigrate(&model.Setting{})
	middleware.DBConn.AutoMigrate(&model.PasswordResetCode{})

}

func main() {
	app := fiber.New(fiber.Config{
		AppName: middleware.GetEnv("PROJ_NAME"),
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow all origins (use specific origins in production)
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// ✅ Define API Routes AFTER CORS
	routes.AppRoutes(app)

	// Enable logger middleware
	app.Use(logger.New())

	middleware.StartOverdueScheduler()

	go controller.StartPenaltyChecker()
	go controller.StartReservationChecker()

	notifications.StartOverdueCheckerCron()
	notifications.InitializeEmailConfig()
	if err := notifications.TestEmailConfig(); err != nil {
		fmt.Printf("❌ Email configuration error: %v\n", err)
		// Don't exit, but log the error
	} else {
		fmt.Println("✅ Email configuration is working!")
	}

	// Start the server
	app.Listen(fmt.Sprintf("0.0.0.0:%s", middleware.GetEnv("PROJ_PORT")))

}
