package shell

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

// canRetryWrite reports whether the write failed because another process
// temporarily holds the file (or its rename target) open, making a retry
// worthwhile. All other errors are considered persistent.
func canRetryWrite(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}

	return errno == windows.ERROR_SHARING_VIOLATION || errno == windows.ERROR_LOCK_VIOLATION
}
