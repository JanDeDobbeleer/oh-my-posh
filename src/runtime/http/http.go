package http

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func dialer() *net.Dialer {
	return &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, network, "8.8.8.8:53") // Use Google DNS
			},
		},
		Timeout: 30 * time.Second,
	}
}

var (
	defaultTransport http.RoundTripper = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer().Dial,
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
