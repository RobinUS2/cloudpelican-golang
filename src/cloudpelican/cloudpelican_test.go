package cloudpelican

// @author Robin Verlangen
// Test the CloudPelican for Go libraby

import "testing"

// Test log message
func TestLogMessage(t *testing.T) {
    const in string = "Hello World"
    const out bool = true
    if x := LogMessage(in); x != out {
        t.Errorf("LogMessage(%v) = %v, want %v", in, x, out)
    }
}
