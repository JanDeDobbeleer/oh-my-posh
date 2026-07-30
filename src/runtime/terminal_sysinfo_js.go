//go:build js

package runtime

// SystemInfo has no CPU load, disk I/O counters, or memory stats to read
// inside a browser - the sysinfo segment writer that is SystemInfo's only
// caller is never wired up in the wasm build (segment writers are nil under
// render.Config's DataOnly path), so this only exists to satisfy the
// Environment interface.
func (term *Terminal) SystemInfo() (*SystemInfo, error) {
	return nil, &NotImplemented{}
}
