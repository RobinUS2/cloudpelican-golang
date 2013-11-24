package cloudpelican

// @author Robin Verlangen
// @todo Support bulk index requests
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
var debugMode = false

// Monitor drain status
var startCounter uint64 = uint64(0)
var startCounterMux sync.Mutex
var doneCounter uint64 = uint64(0)
var doneCounterMux sync.Mutex

// Log queue
var writeAheadBufferSize int = 1000
var writeAhead chan string = make(chan url.Values, writeAheadBufferSize)
var writeAheadInit bool
var dropOnFullWriteAheadBuffer bool = true

// Set token
func SetToken(t string) {
    // Validate before setting
    validateToken(t)
    
    // Store
    TOKEN = t
}

// Set endpoint
func SetEndpoint(e string) {
    // Store
    ENDPOINT = e
}

// Set timeout
func SetBackendTimeout(to time.Duration) {
    backendTimeout = to
}

// Debug
func SetDebugMode(b bool) {
    debugMode = b
}

// Write a message
func LogMessage(msg string) bool {
    // Create fields map
    params := url.Values{}
    params.Add("t", TOKEN)
    params.Add("f[msg]", msg)

    // Push to channel
    return requestAsync(params)
}

// Request async
func requestAsync(params url.Values) bool {
    // Check amount of open items in the channel, if the channel is full, return false and drop this message
    if dropOnFullWriteAheadBuffer {
        var lwa int = len(writeAhead)
        if lwa == writeAheadBufferSize {
            log.Printf("Write ahead buffer is full and contains %d items. Dropping current log message", lwa)
        }
    }

    // Add counter
    startCounterMux.Lock()
    startCounter++
    startCounterMux.Unlock()

    // Do we have to start a writer?
    if writeAheadInit == false {
        writeAheadInit = true
        backendWriter()
    }

    // Insert into channel
    writeAhead <- params

    // OK
    return true
}

// Backend writer
func backendWriter() {
    go func() {
        // Client
        transport := &http.Transport{
            Dial: func(netw, addr string) (net.Conn, error) {
                    deadline := time.Now().Add(backendTimeout)
                    c, err := net.DialTimeout(netw, addr, time.Second)
                    if err != nil {
                            return nil, err
                    }
                    c.SetDeadline(deadline)
                    return c, nil
            }}
        httpclient := &http.Client{Transport: transport}

        // Wait for messages
        for {
            // Read from channel
            var params url.Values
            params = <- writeAhead

            var url string = ENDPOINT + "?" + params.Encode()

            // Make request
            if debugMode {
                log.Printf("Write ahead queue %d\n", len(writeAhead))
                log.Println(url)
            }
            resp, err := httpclient.Get(url)
            if err != nil {
                log.Printf("Error while forwarding data: %s\n", err)
            } else {
                defer resp.Body.Close()
            }

            // Done counter
            doneCounterMux.Lock()
            doneCounter++
            doneCounterMux.Unlock()
        }
        log.Printf("Stopping backend writer")
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