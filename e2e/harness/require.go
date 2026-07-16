package harness

import (
	"os"
	"strings"
	"testing"
)

// requireEnvVar names the environment variable that lists shells the current run must not
// silently skip. See SkipUnavailable.
const requireEnvVar = "OMP_E2E_REQUIRE"

// SkipUnavailable skips the current test with reason, unless shellName is listed in the
// comma-separated OMP_E2E_REQUIRE environment variable, in which case it fails the test
// instead. CI sets OMP_E2E_REQUIRE to the shells it installs, so a shell that's supposed to
// be present but isn't found (or isn't supported on the host) turns into a hard failure
// instead of a quiet skip.
func SkipUnavailable(t *testing.T, shellName, reason string) {
	t.Helper()

	if required(shellName) {
		t.Fatalf("shell %s is required by %s but unavailable: %s", shellName, requireEnvVar, reason)
	}

	t.Skip(reason)
}

// required reports whether shellName (case-insensitive) appears in the comma-separated
// OMP_E2E_REQUIRE environment variable.
func required(shellName string) bool {
	list := os.Getenv(requireEnvVar)
	if list == "" {
		return false
	}

	for name := range strings.SplitSeq(list, ",") {
		if strings.EqualFold(strings.TrimSpace(name), shellName) {
			return true
		}
	}

	return false
}
