package middleware

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env") // Load .env only once
	if err != nil {
		log.Println("Warning: No .env file found, using system env variables")
	}
}

func GetEnv(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading .env file")
		log.Fatalf("Error loading .env file")
		return err.Error()
	}
	return os.Getenv(key)
}
