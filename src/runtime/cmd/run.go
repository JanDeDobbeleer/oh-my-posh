package cmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
)

// Run executes a command in its own process group and returns its output, falling back to stderr
// when stdout is empty. Callers can clean up the process group with KillGoroutineChildren.
func Run(command string, args ...string) (string, error) {
	return RunWithEnv(command, nil, args...)
}

// RunWithEnv executes a command with additional environment variables and returns its output,
// falling back to stderr when stdout is empty.
func RunWithEnv(command string, envs []string, args ...string) (string, error) {
	return runWithEnv(command, envs, true, args...)
}

// RunWithEnvNoFallback executes a command with additional environment variables and returns only
// its stdout when the command succeeds. On error, it returns stderr when available.
func RunWithEnvNoFallback(command string, envs []string, args ...string) (string, error) {
	return runWithEnv(command, envs, false, args...)
}

// runWithEnv executes a command and optionally falls back to stderr when stdout is empty.
func runWithEnv(command string, envs []string, fallbackToStderr bool, args ...string) (string, error) {
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
	if result == "" && fallbackToStderr {
		result = strings.TrimSpace(errb.String())
	}
	return result, nil
}
