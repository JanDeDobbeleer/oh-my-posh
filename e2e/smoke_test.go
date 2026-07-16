package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/e2e/harness"
)

// forbiddenScreenText lists substrings that must never appear on a healthy shell's
// rendered screen right after the prompt comes up: shell-not-found errors, missing-command
// errors, or the init script raising an error of its own. Each phrase is a specific failure
// wording rather than a bare "error", so a benign occurrence (a hostname, a banner) can't
// fail the test.
var forbiddenScreenText = []string{
	"command not found",
	"not recognized",
	"unable to find",
	"syntax error",
	"parse error",
	"failed to",
}

// aliveMarker is echoed back by every shell after the prompt is confirmed, proving the
// session actually accepts and runs interactive input.
const aliveMarker = "omp-e2e-alive"

var aliveMarkerRegexp = regexp.MustCompile(regexp.QuoteMeta(aliveMarker))

// TestSmoke boots every shell interactively with the base config, in a real pty, and
// asserts: the prompt renders with exit code 0, the screen carries none of the forbidden
// error strings, a typed command echoes back, and the shell exits cleanly. Shells whose
// binary is missing or that are unsupported on this platform (see ShellDef.
// SupportedOnHost) skip cleanly via harness.Start.
func TestSmoke(t *testing.T) {
	for _, sh := range harness.Shells {
		t.Run(sh.Name, func(t *testing.T) {
			cfgPath := harness.WriteConfig(t)
			session := harness.Start(t, sh, cfgPath)

			screen := session.WaitForPrompt()

			if !strings.Contains(screen, "E2E:0>") {
				t.Fatalf("%s: prompt screen missing \"E2E:0>\":\n%s", sh.Name, screen)
			}

			lower := strings.ToLower(screen)
			for _, forbidden := range forbiddenScreenText {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("%s: prompt screen contains forbidden text %q:\n%s", sh.Name, forbidden, screen)
				}
			}

			session.SendLine("echo " + aliveMarker)
			session.WaitFor(aliveMarkerRegexp)

			session.ExpectExit()
		})
	}
}
