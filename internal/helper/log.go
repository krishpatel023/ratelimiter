package helper

import (
	"fmt"
	"time"
)

func Log(message, status string) {
	// Log the message
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	status = fmt.Sprintf("[%s]", status)

	fmt.Printf("%s %-10s %s\n", currentTime, status, message)
}
