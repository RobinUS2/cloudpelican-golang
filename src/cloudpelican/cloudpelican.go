package cloudpelican

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "net/http"
    "log"
)

// Settings
const ENDPOINT string = "https://app.cloudpelican.com/api/push/pixel"

// Write a message
func LogMessage(token string, msg string) bool {
    // @todo Write seperate func for url assembly, encoding, etc
    // @todo Validate
    var res bool = requestAsync(ENDPOINT + "?t=" + token + "&f[msg]=" + msg)
    return res
}

// Request a sync
// @todo Make sure all are pushed to backend before application shuts down
func requestAsync(url string) bool {
    go func() {
        _, err := http.Get(url)
        log.Println(url)
	if err != nil {
            log.Printf("Error while forwarding data: %s\n", err)
        }
    }()
    return true
}
