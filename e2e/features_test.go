package e2e

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hinshun/vt10x"
	"github.com/jandedobbeleer/oh-my-posh/e2e/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// featurePtyCols mirrors the pty width the harness fixes every session to (see
// harness/session.go), used here to assert how close to the right edge the rprompt
// renders.
const featurePtyCols = 120

// scenario is one behavior-layer case: the config overlays needed to enable it, any
// shells that must be skipped (declaratively, with a reason), and the assertions to run
// once the session's first prompt is up.
type scenario struct {
	name     string
	overlays []harness.Overlay
	skips    map[string]string // shell name -> skip reason
	run      func(t *testing.T, sh harness.ShellDef, s *harness.Session)
}

// featureScenarios is the behavior-layer test matrix, crossed with harness.Shells by
// TestFeatures.
var featureScenarios = []scenario{
	{
		// exitCode boots with the base config, runs the shell's Fail command, and
		// asserts the next prompt reports the expected exit code via the
		// "E2E:<code>>" template.
		name: "exit-code",
		run: func(t *testing.T, sh harness.ShellDef, s *harness.Session) {
			require.NotEmpty(t, sh.Fail.Command, "%s: no Fail command configured", sh.Name)

			s.SendLine(sh.Fail.Command)

			expected := fmt.Sprintf("E2E:%d>", sh.Fail.Code)
			s.WaitFor(regexp.MustCompile(regexp.QuoteMeta(expected)))
		},
	},
	{
		// transient boots with the transient_prompt overlay, types a command, and
		// asserts that once the command's output and the following prompt have both
		// landed, the screen row that carried the accepted command line was
		// rewritten to start with "TR>" — i.e. the primary prompt on that line was
		// replaced by the transient one.
		//
		// bash only supports the transient prompt inside a ble.sh session (gated on
		// the BLE_SESSION_ID environment variable, see src/shell/bash.go), which this
		// harness's plain `bash --noprofile --rcfile ... -i` session does not
		// provide, so bash is skipped explicitly rather than asserted against.
		name:     "transient",
		overlays: []harness.Overlay{harness.Transient},
		skips: map[string]string{
			"bash": "bash only supports a transient prompt inside a ble.sh session; not supported by this harness",
		},
		run: func(t *testing.T, sh harness.ShellDef, s *harness.Session) {
			const echoedText = "transient-check"

			s.SendLine("echo " + echoedText)

			// (?s) lets '.' cross line boundaries: this only matches once the
			// echoed output AND a subsequent primary prompt are both on screen,
			// proving the command has fully run and the shell moved on to its
			// next prompt.
			readyRe := regexp.MustCompile(`(?s)` + echoedText + `.*E2E:\d+>`)
			screen := s.WaitFor(readyRe)

			var commandLine string
			for _, line := range s.ScreenLines() {
				if strings.Contains(line, "echo "+echoedText) {
					commandLine = line
					break
				}
			}

			require.NotEmpty(t, commandLine,
				"%s: could not find the accepted command line on screen:\n%s", sh.Name, screen)

			trimmed := strings.TrimLeft(commandLine, " ")
			assert.True(t, strings.HasPrefix(trimmed, "TR>"),
				"%s: command line was not rewritten with the transient prompt: %q", sh.Name, commandLine)
		},
	},
	{
		// rprompt boots with the rprompt overlay and asserts the fixed "RMARK"
		// marker renders on the same screen row as the primary "E2E:0>" prompt,
		// right-aligned near the pty's 120th column.
		//
		// bash only renders a right prompt inside a ble.sh session (gated on the
		// BLE_SESSION_ID environment variable, see src/shell/bash.go), which this
		// harness's plain `bash --noprofile --rcfile ... -i` session does not
		// provide, so bash is skipped explicitly rather than asserted against.
		name:     "rprompt",
		overlays: []harness.Overlay{harness.RPrompt},
		skips: map[string]string{
			"bash": "bash only renders a right prompt inside a ble.sh session; not supported by this harness",
		},
		run: func(t *testing.T, sh harness.ShellDef, s *harness.Session) {
			var promptLine string
			for _, line := range s.ScreenLines() {
				if strings.Contains(line, "E2E:0>") {
					promptLine = line
					break
				}
			}

			require.NotEmpty(t, promptLine,
				"%s: could not find the primary prompt row on screen:\n%s", sh.Name, s.Screen())

			require.Contains(t, promptLine, "RMARK",
				"%s: RMARK not found on the same row as the primary prompt: %q", sh.Name, promptLine)

			trimmed := strings.TrimRight(promptLine, " ")
			assert.True(t, strings.HasSuffix(trimmed, "RMARK"),
				"%s: RMARK is not right-aligned, trimmed row does not end with it: %q", sh.Name, promptLine)

			endCol := strings.LastIndex(promptLine, "RMARK") + len("RMARK")
			assert.InDelta(t, featurePtyCols, endCol, 2,
				"%s: RMARK does not end near column %d (ended at %d): %q", sh.Name, featurePtyCols, endCol, promptLine)
		},
	},
	{
		// color boots with the Colored overlay and asserts the fixed "CLR" marker
		// renders with its configured truecolor foreground/background, verifying
		// screen-level color rendering rather than just the raw SGR bytes.
		name:     "color",
		overlays: []harness.Overlay{harness.Colored},
		run: func(t *testing.T, sh harness.ShellDef, s *harness.Session) {
			s.WaitFor(regexp.MustCompile(regexp.QuoteMeta("CLR")))

			fg, bg, found := s.MarkerColor("CLR")
			require.True(t, found, "%s: CLR marker not found on screen:\n%s", sh.Name, s.Screen())

			assert.Equal(t, vt10x.Color(0xff0000), fg,
				"%s: unexpected foreground color for CLR: %#06x", sh.Name, uint32(fg))
			assert.Equal(t, vt10x.Color(0x0000ff), bg,
				"%s: unexpected background color for CLR: %#06x", sh.Name, uint32(bg))
		},
	},
	{
		// ftcs boots with the shell_integration overlay and asserts all four FTCS
		// (Final Term Control Sequence) marks land in the raw byte stream: prompt
		// start (133;A) and command start (133;B) around the first prompt, then
		// pre-execution (133;C) and command-finished (133;D) around a typed command.
		//
		// nu's init script intentionally emits none of the FTCS marks: Features().Nu()'s
		// switch does list a case for FTCSMarks (grouped with several other features), but
		// that case deliberately returns an empty Code (see src/shell/nu.go), so the
		// generated script never gets the hook that prints them.
		name:     "ftcs",
		overlays: []harness.Overlay{harness.ShellIntegration},
		skips: map[string]string{
			"nu": "nu's FTCSMarks case in src/shell/nu.go deliberately emits nothing, so shell_integration marks never appear for nu",
		},
		run: func(t *testing.T, sh harness.ShellDef, s *harness.Session) {
			raw := s.Raw()
			require.Contains(t, raw, "\x1b]133;A",
				"%s: missing FTCS prompt-start mark (133;A) after first prompt:\n%s", sh.Name, raw)
			require.Contains(t, raw, "\x1b]133;B",
				"%s: missing FTCS command-start mark (133;B) after first prompt:\n%s", sh.Name, raw)

			s.SendLine("echo ftcs-check")

			readyRe := regexp.MustCompile(`(?s)ftcs-check.*E2E:\d+>`)
			s.WaitFor(readyRe)

			raw = s.Raw()
			assert.Contains(t, raw, "\x1b]133;C",
				"%s: missing FTCS pre-execution mark (133;C):\n%s", sh.Name, raw)
			assert.Contains(t, raw, "\x1b]133;D",
				"%s: missing FTCS command-finished mark (133;D):\n%s", sh.Name, raw)
		},
	},
}

// TestFeatures runs every featureScenarios case against every harness.Shells entry, as
// "TestFeatures/<scenario>/<shell>". For each shell it skips cleanly (with the scenario's
// declared reason) when the feature is unsupported by this harness, otherwise it writes
// the scenario's config overlays, starts the shell, waits for the first prompt, and hands
// off to the scenario's assertions.
func TestFeatures(t *testing.T) {
	for _, sc := range featureScenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, sh := range harness.Shells {
				t.Run(sh.Name, func(t *testing.T) {
					if reason, skip := sc.skips[sh.Name]; skip {
						t.Skip(reason)
					}

					cfgPath := harness.WriteConfig(t, sc.overlays...)
					session := harness.Start(t, sh, cfgPath)
					session.WaitForPrompt()

					sc.run(t, sh, session)
				})
			}
		})
	}
}
