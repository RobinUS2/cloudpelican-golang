package cloudpelican

// @author Robin Verlangen
// Tool for logging data to CloudPelican directly from Go

// Imports
import (
    "net"
    "net/http"
    "net/url"
    "log"
    "sync"
    "time"
)

// Settings
var ENDPOINT string = "https://app.cloudpelican.com/api/push/pixel"
var TOKEN string = ""
var backendTimeout = time.Duration(5 * time.Second)

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
    validateToken(t)
    
    // Store
    TOKEN = t
}

// Write a message
func LogMessageWithToken(t string, msg string) bool {
    // Create fields map
    var fields map[string]string = make(map[string]string)
    fields["msg"] = msg

    // Push to channel
    return requestAsync(assembleUrl(t, fields))
}

// Write a message
func LogMessage(msg string) bool {
    // Create fields map
    var fields map[string]string = make(map[string]string)
    fields["msg"] = msg

    // Push to channel
    return requestAsync(assembleUrl(TOKEN, fields))
}

// Drain: wait for all data pushes to finish
func Drain() bool {
    if startCounter > doneCounter {
        <-routineQuit
    }
    return true
}

// Assemble url
// @return string Url based on the input fields
func assembleUrl(t string, fields map[string]string) string {
    // Token check
    validateToken(t)

    // Baisc query params
    params := url.Values{}
    params.Add("t", t)

    // Fields
    for k, _ := range fields {
        if len(k) == 0 || len(fields[k]) == 0 {
            log.Printf("Skipping invalid field %s with value %s", k, fields[k])
            continue
        }
        params.Add("f[" + k + "]", fields[k])
    }

    // Final url
    return ENDPOINT + "?" + params.Encode()
}

// Request async
func requestAsync(url string) bool {
    // @todo Check amount of open items in the channel, if the channel is full, return false and drop this message

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
        // Transport
        transport := http.Transport{
            Dial: dialTimeout,
        }

        // Client
        client := http.Client{
            Transport: &transport,
        }
        
        // Wait for messages
        for {
            // Read from channel
            var url string
            url = <- writeAhead

            // Make request
            _, err := client.Get(url)
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

// Timeout helper
func dialTimeout(network, addr string) (net.Conn, error) {
    return net.DialTimeout(network, addr, backendTimeout)
}

// Validate the token
func validateToken(t string) {
    if len(t) == 0 {
        log.Println("Please set a valid token with cloudpelican.SetToken(token string)")
    }
}