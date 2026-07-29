//go:build !windows && !js

package runtime

import (
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Root is split out from terminal_unix.go for the same reason
// terminal_writable_unix.go's DirIsWritable is: os.Geteuid() does compile
// for js/wasm (syscall's js implementation forwards it to a JS "process"
// object - see syscall_js.go), but there is no real uid to ask for outside
// an actual OS process, and calling it panics the moment there is no
// "process" global to forward to, which is exactly the case in a browser
// (and in a plain Node harness that never sets one up - see the js variant's
// own doc comment). Splitting this one method behind the narrower tag keeps
// the rest of this file's Terminal implementation - which config and
// segments need regardless of platform - building for js/wasm too.
func (term *Terminal) Root() bool {
	defer log.Trace(time.Now())
	return os.Geteuid() == 0
}
