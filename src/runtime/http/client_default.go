//go:build !(js && wasm)

//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
package http

import (
	"net"
	"net/http"
	"time"
)

var defaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout: 10 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
}

func init() {
	HTTPClient = &http.Client{Transport: defaultTransport}
}
