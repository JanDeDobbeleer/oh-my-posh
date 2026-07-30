//go:build !js

package runtime

import (
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/log"

	process "github.com/shirou/gopsutil/v4/process"
)

// shellProcessName asks the OS for the parent process' name - the fallback
// Shell() uses when no --shell flag was passed. Isolated here so the js/wasm
// build (terminal_shell_js.go), which has no parent process to query, never
// links gopsutil/process.
func (term *Terminal) shellProcessName() string {
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))

	name, err := p.Name()
	if err != nil {
		log.Error(err)
		return ""
	}

	return name
}
