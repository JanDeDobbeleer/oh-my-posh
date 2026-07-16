// Package harness provides shared helpers for the oh-my-posh shell end-to-end tests:
// building the omp binary once per test run, generating configs and init scripts, and
// (in later layers) driving interactive shell sessions in a pty.
package harness

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// errCallerInfo is returned when the Go runtime cannot report the location of this
// source file, which should never happen in practice.
var errCallerInfo = errors.New("could not determine source file location")

// srcDir resolves ../src relative to this source file, not the test binary's working
// directory, so Binary works no matter which package's test invokes it.
func srcDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errCallerInfo
	}

	// this file lives in e2e/harness/binary.go, so src is two levels up.
	return filepath.Abs(filepath.Join(filepath.Dir(file), "..", "..", "src"))
}

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// Binary returns the absolute path to the oh-my-posh executable used by the e2e tests.
// It builds ../src once per test run (guarded by sync.Once) into a temp directory. Set
// OMP_E2E_BINARY to skip the build and use a prebuilt binary instead. The test is failed
// with the build output if the build fails.
func Binary(t *testing.T) string {
	t.Helper()

	if override := os.Getenv("OMP_E2E_BINARY"); override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			t.Fatalf("resolving OMP_E2E_BINARY %q: %v", override, err)
		}

		return abs
	}

	buildOnce.Do(func() {
		binPath, buildErr = buildBinary()
	})

	if buildErr != nil {
		t.Fatalf("building omp binary: %v", buildErr)
	}

	return binPath
}

func buildBinary() (string, error) {
	outDir, err := os.MkdirTemp("", "omp-e2e-bin")
	if err != nil {
		return "", err
	}

	name := "oh-my-posh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	out := filepath.Join(outDir, name)

	src, err := srcDir()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = src

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &buildError{output: string(output), err: err}
	}

	return out, nil
}

// buildError wraps a failed build's combined output so callers can surface it verbatim.
type buildError struct {
	err    error
	output string
}

func (e *buildError) Error() string {
	return e.err.Error() + "\n" + e.output
}

func (e *buildError) Unwrap() error {
	return e.err
}
