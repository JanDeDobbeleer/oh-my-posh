// Package e2e contains layer 1 (syntax), layer 2 (smoke) and layer 3 (behavior)
// end-to-end tests for the oh-my-posh shell integrations.
package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/e2e/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// placeholders lists template tokens that must never survive into a generated init
// script; their presence means rendering silently failed to substitute a value.
var placeholders = []string{"::OMP::", "::CONFIG::", "::SESSION_ID::"}

// overlaySets is the feature-config matrix layer 1 runs against every shell: the base
// config alone, and each documented overlay (or combination) applied on top of it.
var overlaySets = []struct {
	Name     string
	Overlays []harness.Overlay
}{
	{Name: "base"},
	{Name: "transient", Overlays: []harness.Overlay{harness.Transient}},
	{Name: "rprompt", Overlays: []harness.Overlay{harness.RPrompt}},
	{Name: "tooltips", Overlays: []harness.Overlay{harness.Tooltips}},
	{Name: "shell_integration", Overlays: []harness.Overlay{harness.ShellIntegration}},
	{Name: "full", Overlays: []harness.Overlay{harness.Full}},
}

// TestSyntax generates the init script for every shell x config-overlay combination and
// validates it with the shell's own parser. The non-empty/no-placeholder assertions run
// unconditionally; the parser check only runs when the shell binary is available and
// skips cleanly otherwise.
func TestSyntax(t *testing.T) {
	for _, sh := range harness.Shells {
		for _, overlay := range overlaySets {
			t.Run(sh.Name+"/"+overlay.Name, func(t *testing.T) {
				cfgPath := harness.WriteConfig(t, overlay.Overlays...)
				script := harness.InitScript(t, sh.Name, cfgPath)

				require.NotEmpty(t, script, "init script for %s must not be empty", sh.Name)

				for _, placeholder := range placeholders {
					assert.NotContains(t, script, placeholder, "leftover placeholder in %s init script", sh.Name)
				}

				if _, err := harness.LookupShellBinary(sh.Binary); err != nil {
					harness.SkipUnavailable(t, sh.Name, fmt.Sprintf("%s not found on PATH, skipping syntax check", sh.Binary))
				}

				scriptPath := harness.WriteScript(t, sh.Name, script)

				cmd := sh.SyntaxCheck(scriptPath)

				output, err := cmd.CombinedOutput()
				require.NoErrorf(t, err, "syntax check failed for %s:\n%s", sh.Name, output)
			})
		}
	}
}

// TestTransientPromptOverlay verifies the transient_prompt overlay's shape against a real
// omp invocation: config.Config.TransientPrompt is a *Segment, so the overlay must marshal
// to an object (not a string), and enabling it must actually change the generated pwsh
// script versus the base config.
func TestTransientPromptOverlay(t *testing.T) {
	baseCfg := harness.WriteConfig(t)
	transientCfg := harness.WriteConfig(t, harness.Transient)

	baseScript := harness.InitScript(t, "pwsh", baseCfg)
	transientScript := harness.InitScript(t, "pwsh", transientCfg)

	require.NotEmpty(t, baseScript)
	require.NotEmpty(t, transientScript)

	assert.NotEqual(t, baseScript, transientScript, "transient overlay should change the generated script")

	const transientMarker = "$global:_ompTransientPrompt = $true"
	assert.True(t, strings.Contains(transientScript, transientMarker), "transient script missing %q", transientMarker)
	assert.False(t, strings.Contains(baseScript, transientMarker), "base script unexpectedly contains %q", transientMarker)
}
