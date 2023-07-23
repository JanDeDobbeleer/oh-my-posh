package platform

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func Proxy(_ *http.Request) (*url.URL, error) {
	proxyURL := os.Getenv("HTTPS_PROXY")
	if len(proxyURL) == 0 {
		return nil, nil
	}
	return url.Parse(proxyURL)
}

var (
	defaultTransport http.RoundTripper = &http.Transport{
		Proxy: Proxy,
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	Client httpClient = &http.Client{Transport: defaultTransport}
)
