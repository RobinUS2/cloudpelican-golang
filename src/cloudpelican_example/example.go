package main

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "cloudpelican"
    "log"
)

// Example
func main() {
    var msg string = "This is a log message"

    // Write message and validate
    res := cloudpelican.LogMessage("12345", msg)
    if !res {
        log.Println("Something went wrong")
    } else {
        log.Printf("Written data to backend: %s\n", msg)
    }
}
