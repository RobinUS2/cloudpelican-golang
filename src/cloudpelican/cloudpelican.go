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

// Monitor drain status
routineQuit chan int = make(chan int)
startCounter uint64 = uint64(0)
startCounterMux sync.Mutex
doneCounter uint64 = uint64(0)
doneCounterMux sync.Mutex

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
}

// Request a sync
// @todo Make sure all are pushed to backend before application shuts down
func requestAsync(url string) bool {
    // Add counter
    startCounterMux.Lock()
    startcounter++
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
        doneCounter.Unlock()
        // @todo Check whether dif between started and done is = 0, if so, drop a message in the routineQuit
    }()
    return true
}
