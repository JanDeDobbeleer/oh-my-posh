//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
package http

import (
	"context"
	"net"
	"time"
)

// IsConnected checks if we can connect to ohmyposh within 200ms.
// Exposed as a variable so it can be replaced in tests.
var IsConnected = func() bool {
	timeout := 200 * time.Millisecond
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	ctx := context.Background()
	conn, err := dialer.DialContext(ctx, "tcp", "ohmyposh.dev:80")
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
