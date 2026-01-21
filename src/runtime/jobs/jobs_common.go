package jobs

import (
	"runtime"
	"strconv"
	"strings"
)

// CurrentGID returns the current goroutine's id. We expose this here so
// callers can register PIDs without parsing runtime.Stack in multiple
// places.
func CurrentGID() uint64 {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	s := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))
	if len(s) == 0 {
		return 0
	}
	idStr := s[0]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
