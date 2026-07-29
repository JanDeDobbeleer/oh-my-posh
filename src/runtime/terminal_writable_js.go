//go:build js

package runtime

// DirIsWritable always reports false on js/wasm: there is no filesystem
// permission model to query when running inside a browser or other wasm
// host, unlike the real unix.Access check terminal_writable_unix.go
// performs on every other non-Windows platform.
func (term *Terminal) DirIsWritable(_ string) bool {
	return false
}
