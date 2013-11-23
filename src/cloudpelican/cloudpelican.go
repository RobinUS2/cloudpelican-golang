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
var token string = ""

// Monitor drain status
var routineQuit chan int = make(chan int)
var startCounter uint64 = uint64(0)
var startCounterMux sync.Mutex
var doneCounter uint64 = uint64(0)
var doneCounterMux sync.Mutex

// Log queue
var writeAheadBufferSize int = 1000
var writeAhead chan string = make(chan string, writeAheadBufferSize)
var writeAheadInit bool

// Set token
func SetToken(t string) {
    // Validate before setting
    validateToken(token)
    
    // Store
    token = t
}

// Write a message
func LogMessageWithToken(t string, msg string) bool {
    // Token check
    validateToken(t)
    
    // @todo Write seperate func for url assembly, encoding, etc
    // @todo Validate
    var res bool = requestAsync(ENDPOINT + "?t=" + t + "&f[msg]=" + msg)
    return res
}

// Write a message
func LogMessage(msg string) bool {
    // Token check
    validateToken(token)
    
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

// Request async
func requestAsync(url string) bool {
    // Add counter
    startCounterMux.Lock()
    startCounter++
    startCounterMux.Unlock()

    // Insert into channel
    writeAhead <- url

    // Do we have to start a writer?
    if writeAheadInit == false {
        writeAheadInit = true
        backendWriter()
    }

    // OK
    return true
}

// Backend writer
func backendWriter() {
    go func() {
        for {
            // Read from channel
            var url string
            url = <- writeAhead

            // Make request
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
        }
    }()
}

// Validate the token
func validateToken(t string) {
    if len(t) == 0 {
        log.Println("Please set a valid token with cloudpelican.SetToken(token string)")
    }
}