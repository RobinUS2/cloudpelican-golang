package main

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "github.com/RobinUS2/cloudpelican-golang/cloudpelican"
    "log"
    "fmt"
    "time"
)

// Example
func main() {
    // Token
    cloudpelican.SetToken("12345")

    // More verbose
    cloudpelican.SetDebugMode(true)

    time.Sleep(1000 * time.Millisecond)

    // Basic message
    var msg string = "This is a log message %d"

    // Write message and validate
    for i := 0; i < 10; i++ {
        // Basic message with a sequence number for esting purposes
        parsedMsg := fmt.Sprintf(msg, i)
        res := cloudpelican.LogMessage(parsedMsg)
        time.Sleep(1000 * time.Millisecond)
        // Validate the writing
        if !res {
            log.Println("Something went wrong")
        } else {
            log.Printf("Written %d bytes of data '%s' to backend.\n", len(parsedMsg), parsedMsg)
        }
    }
}