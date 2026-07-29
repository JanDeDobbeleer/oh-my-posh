package cmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
)

// Starts the process in its own process group and records it so callers can request
// cleanup via KillGoroutineChildren if they abort waiting for the spawning goroutine.
func Run(command string, args ...string) (string, error) {
	return RunWithEnv(command, nil, args...)
}

func RunWithEnv(command string, envs []string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), command, args...)
	if len(envs) > 0 {
		cmd.Env = append(os.Environ(), envs...)
	}

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
