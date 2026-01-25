//go:build freebsd && (arm || 386)

package daemon

import "syscall"

func setIdent(event *syscall.Kevent_t, pid int) {
	event.Ident = uint32(pid)
}
