package main

import (
	"context"
	"net/http"
)

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	client httpClient = &http.Client{}
)

// Get an HTTP response by sending an HTTP GET request to the specified URL.
func Get(url string, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header = headers
	return client.Do(request)
}
