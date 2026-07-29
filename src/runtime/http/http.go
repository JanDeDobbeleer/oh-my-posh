//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
package http

import (
	"net/http"
)

// Inspired by: https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPClient is what every segment's own request goes through (see
// runtime.Terminal.HTTPRequest). Its value is set per platform: client_default.go builds a real
// transport, client_js.go refuses outright. See client_js.go for why.
var HTTPClient httpClient

type Error struct {
	StatusCode int
}

func (e *Error) Error() string {
	return http.StatusText(e.StatusCode)
}
