package main

import "net/http"

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	getDoFunc func(req *http.Request) (*http.Response, error)
)

// Do is the mock client's `Do` func
func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return getDoFunc(req)
}
