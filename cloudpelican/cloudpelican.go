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
    "strconv"
    "os"
    "fmt"
    "time"
)

// Settings
var TOKEN string = "" // Put your API token here
var ENDPOINT string = "https://api.cloudpelican.com/api"
var backendTimeout = time.Duration(5 * time.Second)
var debugMode bool = false
var maxBulkSize uint64 = uint64(100)
var hostname string

// Monitor drain status
var startCounter uint64 = uint64(0)
var startCounterMux sync.Mutex
var doneCounter uint64 = uint64(0)
var doneCounterMux sync.Mutex
var isDraining bool = false
var drained = make(chan bool); 

// Log queue
var writeAheadBufferSize int = 1000
var writeAhead chan map[string]string = make(chan map[string]string, writeAheadBufferSize)
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

// Drain
func Drain() {
    isDraining = true
    if startCounter > doneCounter {
        // Wait for signal
        <- drained
    }
}

// Write a message
func LogMessage(msg string) bool {
    // Create fields map
    params := make(map[string]string)
    params["__token__"] = TOKEN
    params["msg"] = msg
    params["dt"] = getTimeString()

    // Push to channel
    return requestAsync(params)
}

// Current time
func getTimeString() string {
    return fmt.Sprintf("%d", int64(time.Now().Unix()) * int64(1000))
}

// Request async
func requestAsync(params map[string]string) bool {
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

// Get hostname of this system
func getHostname() string {
    // Hostname
    name, err := os.Hostname()
    if err != nil {
        return ""
    }
    return name 
}

// Backend writer
func backendWriter() {
    hostname = getHostname()
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
        var urlParams url.Values
        var currentEventCount uint64 = uint64(0)
        for {
            // Read from channel
            var fields map[string]string
            fields = <- writeAhead

            // Populate url params
            if currentEventCount == 0 {
                urlParams = url.Values{}
            }
            for k, _ := range fields {
                if k == "__token__" {
                    // Token
                    urlParams.Add("t", fields[k]);
                } else {
                    // Field
                    urlParams.Add("f[" + strconv.FormatUint(currentEventCount, 10) + "][" + k + "]", fields[k]);
                }
            }

            // Host
            if len(hostname) > 0 {
                urlParams.Add("f[" + strconv.FormatUint(currentEventCount, 10) + "][host]", hostname);
            }

            // Increase current count
            currentEventCount++

            // Queue length
            var qLen = len(writeAhead)
            if qLen > 0 && currentEventCount < maxBulkSize {
                // There is more in the current queue, bulk request
                continue
            }

            // Assemble url
            var url string = ENDPOINT + "/push/bulk"

            // Make request
            if debugMode {
                log.Printf("Write ahead queue %d\n", qLen)
                log.Println(urlParams.Encode())
            }
            resp, err := httpclient.PostForm(url, urlParams)
            if err != nil {
                log.Printf("Error while forwarding data: %s\n", err)
            } else {
                defer resp.Body.Close()
            }

            // Done counter
            doneCounterMux.Lock()
            doneCounter += currentEventCount
            doneCounterMux.Unlock()

            // Reset event count
            currentEventCount = 0

            // Are we draining the system?
            if isDraining && doneCounter >= startCounter {
                // Flag the drained channel
                drained <- true
            }
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
