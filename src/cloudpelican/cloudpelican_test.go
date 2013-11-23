package cloudpelican

// @author Robin Verlangen
// Test the CloudPelican for Go libraby

import "testing"
import "cloudpelican"

// Test log message
func TestLogMessage(t *testing.T) {
    const in string = "Hello World"
    const out bool = true
    if x := cloudpelican.LogMessage(in); x != out {
        t.Errorf("cloudpelican.LogMessage(%v) = %v, want %v", in, x, out)
    }
}

// Test log message with token
func TestLogMessageWithToken(t *testing.T) {
    const in string = "Hello World"
    const token string = "54321"
    const out bool = true
    if x := cloudpelican.LogMessageWithToken(token, in); x != out {
        t.Errorf("cloudpelican.LogMessageWithToken(%v) = %v, want %v", in, x, out)
    }
}