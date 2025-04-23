package middleware

import (
    "log"
)

// LogError logs error messages to the console or a file
func LogError(message string) {
    log.Println("ERROR:", message)
}