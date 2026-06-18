package cmd

import (
	"fmt"
	"os"
	"testing"

	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
)

func TestCurrentGID(t *testing.T) {
	if gid := runjobs.CurrentGID(); gid == 0 {
		t.Fatalf("CurrentGID returned 0")
	}
}

func TestRunWithEnv(t *testing.T) {
	t.Setenv("OMP_TEST_CHILD_ENV", "parent")

	output, err := RunWithEnv(
		os.Args[0],
		[]string{
			"GO_WANT_HELPER_PROCESS=1",
			"OMP_TEST_CHILD_ENV=injected",
		},
		"-test.run=TestRunWithEnvHelperProcess",
	)
	if err != nil {
		t.Fatalf("RunWithEnv returned error: %v", err)
	}

	if output != "injected" {
		t.Fatalf("RunWithEnv output = %q, want %q", output, "injected")
	}
}

func TestRunWithEnvHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprint(os.Stdout, os.Getenv("OMP_TEST_CHILD_ENV"))
	os.Exit(0)
}
