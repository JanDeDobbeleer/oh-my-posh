//go:build js && wasm

//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
package http

import (
	"errors"
	"net/http"
)

// The wasm build renders a config from recorded data and never reaches the network: every caller
// of this client goes through runtime.Terminal.HTTPRequest, which already refuses when
// runtime.Flags.DataOnly is set - the only mode the browser entrypoint (src/wasm/main.go) runs in.
//
// Constructing a real http.Transport anyway is what pulls the whole transport, TLS, HTTP/2 and DNS
// stack into the module: measured at 3.26 MB raw, 0.46 MB brotli, for code no visitor can execute.
// A stub that refuses keeps the package's API identical - every segment still compiles unchanged -
// while letting the linker drop all of it.
var errNoNetwork = errors.New("the wasm build renders from data only and cannot make requests")

type stubClient struct{}

func (stubClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, errNoNetwork
}

func init() {
	HTTPClient = stubClient{}
}
