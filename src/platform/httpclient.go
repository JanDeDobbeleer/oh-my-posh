package platform

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
	Client httpClient = &http.Client{Transport: defaultTransport}
)
