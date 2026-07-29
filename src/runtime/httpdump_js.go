//go:build js && wasm

package runtime

import (
	"net/http"
)

// The wasm build never issues a request: HTTPRequest refuses before reaching here whenever
// DataOnly is set, which is the only mode the browser entrypoint runs in (see
// runtime/http/client_js.go). Importing net/http/httputil purely to format a dump nothing can
// produce costs ~106 KB brotli, so this build formats nothing.
func dumpRequest(_ *http.Request) string {
	return ""
}
