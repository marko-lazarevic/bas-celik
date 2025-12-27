// Package logger provides simple logging functionalities for the application.
package logger

import (
	"io"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// DisableLog disables all logging output
func DisableLog() {
	log.SetOutput(io.Discard)
}

// Error logs error messages
func Error(err error) {
	log.Println("ERROR", err.Error())
}

// Info logs informational messages
func Info(message string) {
	log.Println("INFO", message)
}

// Debug logs debug messages when debug mode is enabled
func Debug(message string) {
	log.Println("DEBUG", message)
}
