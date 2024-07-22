package http

import (
	"net"
	"time"
)

// if we can connect to google within 200ms, we are connected
// otherwise, let's consider being offline
func IsConnected() bool {
	timeout := 200 * time.Millisecond
	_, err := net.DialTimeout("tcp", "google.com:80", timeout)
	return err == nil
}
