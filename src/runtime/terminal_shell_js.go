//go:build js

package runtime

// shellProcessName has no parent process to query inside a browser (or a
// Node test harness with no "process" global) - Shell() already falls back to
// UNKNOWN when this returns empty, so there is nothing this can meaningfully
// answer here, unlike terminal_shell.go's real os.Getppid + gopsutil/process
// lookup on every other platform.
func (term *Terminal) shellProcessName() string {
	return ""
}
