package cloudpelican

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "net/http"
    "log"
)

// Write a message
func LogMessage(token string, msg string) bool {
    // @todo
    return false
}

// Request a sync
func requestAsync(url string) bool {
    go func() {
        _, err := http.Get(url)
	if err != nil {
            log.Printf("Error while forwarding data: %s\n", err)
        }
    }()
    return true
}
