package main

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "cloudpelican"
    "log"
    "strconv"
)

// Token
const TOKEN string = "123456"

// Example
func main() {
    var msg string = "This is a log message"

    // Write message and validate
    for i := 0; i < 10; i++ {
        res := cloudpelican.LogMessage(TOKEN, msg + " " + strconv.Itoa(i))
        if !res {
            log.Println("Something went wrong")
        } else {
            log.Printf("Written %d bytes of data '%s' to backend.\n", len(msg), msg)
        }
    }

    // Make sure any pending messages are written to the backend
    cloudpelican.Drain()
}