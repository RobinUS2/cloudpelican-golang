package cloudpelican

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "net/http"
    "log"
    "sync"
)

// Settings
const ENDPOINT string = "https://app.cloudpelican.com/api/push/pixel"

// Monitor drain status
var routineQuit chan int = make(chan int)
var startCounter uint64 = uint64(0)
var startCounterMux sync.Mutex
var doneCounter uint64 = uint64(0)
var doneCounterMux sync.Mutex

// Write a message
func LogMessage(token string, msg string) bool {
    // @todo Write seperate func for url assembly, encoding, etc
    // @todo Validate
    var res bool = requestAsync(ENDPOINT + "?t=" + token + "&f[msg]=" + msg)
    return res
}

// Drain: wait for all data pushes to finish
func Drain() bool {
    if startCounter > doneCounter {
        <-routineQuit
    }
    return true
}

// Request a sync
// @todo Make sure all are pushed to backend before application shuts down
func requestAsync(url string) bool {
    // Add counter
    startCounterMux.Lock()
    startCounter++
    startCounterMux.Unlock()

    // Launch
    go func() {
        // Request url
        _, err := http.Get(url)
        log.Println(url)
	if err != nil {
            log.Printf("Error while forwarding data: %s\n", err)
        }

        // Done counter
        doneCounterMux.Lock()
        doneCounter++
        doneCounterMux.Unlock()

        // Check whether dif between started and done is = 0, if so, drop a message in the routineQuit
        if (doneCounter >= startCounter) {
            routineQuit <- 1
        }
    }()
    return true
}
