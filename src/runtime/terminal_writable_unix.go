//go:build !windows && !js

package runtime

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"golang.org/x/sys/unix"
)

// DirIsWritable is split out from terminal_unix.go because golang.org/x/sys/unix
// does not build for js/wasm: config and segments need the rest of this
// file's Terminal implementation to compile there too, so only the one
// method that actually depends on unix stays behind this narrower tag.
func (term *Terminal) DirIsWritable(input string) bool {
	defer log.Trace(time.Now(), input)
	return unix.Access(input, unix.W_OK) == nil
}
