package cmd

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
)

// Run executes a command while ensuring the OS process is started in its own
// process group; the started process is recorded so callers can request a
// cleanup (KillGoroutineChildren) if they decide to abort waiting for the
// goroutine that spawned it.
func Run(command string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), command, args...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	// ensure child runs in its own process group so we can kill the tree if
	// needed. Implementation is provided by the runtime/jobs package which is
	// platform aware.
	runjobs.SetProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return "", err
	}

	// register the started process under the current goroutine
	runjobs.RegisterProcess(cmd.Process.Pid)
	defer runjobs.UnregisterProcess(cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		// Prefer stderr if available
		output := strings.TrimSpace(errb.String())
		if output == "" {
			output = strings.TrimSpace(out.String())
		}
		return output, err
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		result = strings.TrimSpace(errb.String())
	}
	return result, nil
}
