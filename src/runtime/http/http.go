//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
package http

import (
	"net"
	"net/http"
	"time"
)

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	defaultTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	HTTPClient httpClient = &http.Client{Transport: defaultTransport}
)

type Error struct {
	StatusCode int
}

func (e *Error) Error() string {
	return http.StatusText(e.StatusCode)
}
