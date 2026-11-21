package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/e2e/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNuStreamingUpdates(t *testing.T) {
	tests := []struct {
		name     string
		want     []string
		overlays []harness.Overlay
	}{
		{name: "single-line"},
		{name: "multiline", overlays: []harness.Overlay{harness.MultiLine}, want: []string{"ML1", "ML2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixtureDir := t.TempDir()
			release := filepath.Join(fixtureDir, "release")
			requested := filepath.Join(fixtureDir, "requested")
			fixture := filepath.Join(fixtureDir, "version.nu")
			script := fmt.Sprintf(`
'requested' | save --force '%s'
for _ in 1..300 {
    if ('%s' | path exists) { break }
    sleep 50ms
}
print '1.2.3'
`, filepath.ToSlash(requested), filepath.ToSlash(release))
			require.NoError(t, os.WriteFile(fixture, []byte(script), 0o600))
			t.Cleanup(func() { _ = os.WriteFile(release, nil, 0o600) })

			overlay := func(cfg map[string]any) {
				cfg["streaming"] = 100
				blocks := cfg["blocks"].([]any)
				block := blocks[len(blocks)-1].(map[string]any)
				block["segments"] = append(block["segments"].([]any), map[string]any{
					"type":        "language",
					"style":       "plain",
					"template":    "RESOLVED:{{ .Full }}",
					"placeholder": "PENDING",
					"options": map[string]any{
						"name":         "delayed",
						"display_mode": "always",
						"tools": []any{map[string]any{
							"name":       "delayed",
							"executable": "nu",
							"args":       []string{"--no-config-file", fixture},
							"regex":      `(?P<version>1\.2\.3)`,
						}},
					},
				})
			}
			cfg := harness.WriteConfig(t, append(tc.overlays, overlay)...)
			session := harness.Start(t, nuShell(t), cfg)
			session.WaitFor(regexp.MustCompile(`E2E:\d+>PENDING`))
			require.FileExists(t, requested, "the version command must actually be pending")
			assert.NotContains(t, session.Raw(), "RESOLVED")
			session.SendText("echo preserved-input")
			session.WaitFor(regexp.MustCompile(`PENDING.*echo preserved-input`))
			require.NoError(t, os.WriteFile(release, nil, 0o600))
			screen := session.WaitFor(regexp.MustCompile(`RESOLVED:1\.2\.3.*echo preserved-input`))
			for _, want := range tc.want {
				assert.Contains(t, screen, want)
			}
			assert.NotContains(t, session.Raw(), "Error:")
			require.NoError(t, os.Remove(release))
			session.SendLine("")
			session.WaitFor(regexp.MustCompile(`(?s)preserved-input.*preserved-input.*E2E:\d+>PENDING`))
			session.SendText("echo second-input")
			require.NoError(t, os.WriteFile(release, nil, 0o600))
			session.WaitFor(regexp.MustCompile(`RESOLVED:1\.2\.3.*echo second-input`))
		})
	}
}

func nuShell(t *testing.T) harness.ShellDef {
	t.Helper()

	for _, sh := range harness.Shells {
		if sh.Name == "nu" {
			return sh
		}
	}

	t.Fatal("nu shell definition not found")
	return harness.ShellDef{}
}

func TestNuStreamingDaemonLifecycle(t *testing.T) {
	cfg := harness.WriteConfig(t, func(cfg map[string]any) { cfg["streaming"] = 100 })
	session := harness.Start(t, nuShell(t), cfg)

	session.SendLine("let daemon = $env._OMP_SERVE_JOB; print $daemon")
	session.WaitFor(regexp.MustCompile(`(?s)print \$daemon\s+\d+\s+E2E:\d+>`))

	for cycle := range 3 {
		session.SendLine(fmt.Sprintf("print 'streaming-check-%d'", cycle))
		session.WaitFor(regexp.MustCompile(fmt.Sprintf(`(?s)'streaming-check-%d'\s+streaming-check-%d\s+E2E:\d+>`, cycle, cycle)))
	}

	session.SendLine("print ($daemon == $env._OMP_SERVE_JOB)")
	session.WaitFor(regexp.MustCompile(`(?s)print \(\$daemon == \$env\._OMP_SERVE_JOB\)\s+true\s+E2E:\d+>`))
	assert.NotContains(t, session.Raw(), "Error:")

	session.ExpectExit()
}

func TestNuStreamingTracksExecutedCommands(t *testing.T) {
	cfg := harness.WriteConfig(t, func(cfg map[string]any) { cfg["streaming"] = 100 })
	session := harness.Start(t, nuShell(t), cfg)

	tests := []struct {
		input string
		want  string
	}{
		{input: "print 1", want: "1"},
		{input: "print $env.POSH_EXECUTED", want: "true"},
		{input: "", want: ""},
		{input: "print $env.POSH_EXECUTED", want: "false"},
		{input: "print $env.POSH_EXECUTED", want: "true"},
	}

	for _, tc := range tests {
		session.SendLine(tc.input)
		if tc.want == "" {
			session.WaitForPrompt()
			continue
		}

		session.WaitFor(regexp.MustCompile(fmt.Sprintf(`(?s)%s\s+%s\s+E2E:\d+>`, regexp.QuoteMeta(tc.input), tc.want)))
	}
}
