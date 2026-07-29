//go:build js

package runtime

// Root always reports false on js/wasm: there is no real uid to ask for
// running inside a browser (or a Node test harness with no "process"
// global set up), unlike the real os.Geteuid() check
// terminal_root_unix.go performs on every other non-Windows platform. This
// mirrors DirIsWritable's own js override (terminal_writable_js.go) - "no
// permission model to query here" - for the same underlying reason: an
// os.Geteuid() call reaches syscall's js implementation, which forwards it
// to a JS "process" global (see syscall_js.go's jsProcess) that simply does
// not exist in a browser, and calling it there panics rather than erroring.
func (term *Terminal) Root() bool {
	return false
}
