//go:build !(js && wasm)

package runtime

import (
	"net/http"
	"net/http/httputil"
)

// dumpRequest renders a request for --debug output.
//
// DumpRequest, not DumpRequestOut: the latter shows the exact bytes that would go on the wire,
// which it manages by building an http.Transport internally - a second, hidden transport
// construction that keeps the entire TLS stack linked even where the real client is stubbed out
// (see runtime/http/client_js.go). What it adds over this is the headers the transport itself
// would set: Host, User-Agent, Accept-Encoding. What a segment is actually debugged against - the
// URL, its own headers, the body - is identical either way, and both restore the body after
// reading it.
func dumpRequest(request *http.Request) string {
	dump, _ := httputil.DumpRequest(request, true)
	return string(dump)
}
