package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/e2e/harness"
	"github.com/stretchr/testify/require"
)

func TestSecondaryPromptInitialization(t *testing.T) {
	shells := []struct {
		args   func(string) []string
		name   string
		binary string
		suffix string
	}{
		{
			name: "pwsh", binary: "pwsh",
			suffix: `[Console]::Out.Write((Get-PSReadLineOption).ContinuationPrompt)`,
			args:   func(path string) []string { return []string{"-NoProfile", "-File", path} },
		},
		{
			name: "bash", binary: "bash",
			suffix: `printf '%s' "$_omp_secondary_prompt"`,
			args:   func(path string) []string { return []string{path} },
		},
		{
			name: "zsh", binary: "zsh",
			suffix: `printf '%s|%s' "$_omp_secondary_prompt" "$POSH_MULTILINE_KEEPPROMPT"`,
			args:   func(path string) []string { return []string{path} },
		},
		{
			name: "nu", binary: "nu",
			suffix: `print -n $env.PROMPT_MULTILINE_INDICATOR`,
			args:   func(path string) []string { return []string{"--no-config-file", path} },
		},
		{
			name: "yash", binary: "yash",
			suffix: `eval "printf '%s' \"$YASH_PS2\""`,
			args:   func(path string) []string { return []string{path} },
		},
	}

	cases := []struct {
		name     string
		template string
		custom   bool
	}{
		{name: "default"},
		{name: "custom", template: "{{ .Shell }} {{ .Env.OMP_SECONDARY_TEST }} \" ' \\\n unicode-\u00b2 > ", custom: true},
		{name: "empty", template: "{{ \"\" }}", custom: true},
	}

	for _, sh := range shells {
		for _, tc := range cases {
			t.Run(sh.name+"/"+tc.name, func(t *testing.T) {
				if runtime.GOOS == "windows" && sh.name != "pwsh" && sh.name != "nu" {
					t.Skip("POSIX shell execution requires a Unix host")
				}

				bin, err := harness.LookupShellBinary(sh.binary)
				if err != nil {
					t.Skipf("%s unavailable: %v", sh.binary, err)
				}

				cfg := harness.WriteConfig(t, func(cfg map[string]any) {
					if tc.custom {
						cfg["secondary_prompt"] = map[string]any{
							"foreground":           "#ff0000",
							"template":             tc.template,
							"multiline_keepprompt": true,
						}
					}
				})
				omp := harness.Binary(t)
				cacheDir := t.TempDir()
				t.Setenv("OMP_SECONDARY_TEST", "init")
				env := append(os.Environ(), "OMP_CACHE_DIR="+cacheDir)
				printCommand := exec.CommandContext(t.Context(), omp, "print", "secondary", "--shell="+sh.name, "--config="+cfg)
				printCommand.Env = env
				want, err := printCommand.Output()
				require.NoError(t, err)

				if sh.name == "zsh" {
					want = append(want, []byte("|"+map[bool]string{true: "true", false: "false"}[tc.custom])...)
				}

				modes := []struct {
					name  string
					flags []string
				}{
					{name: "print", flags: []string{"--print"}},
					{name: "cold-cache"},
					{name: "warm-cache"},
				}
				if sh.name == "nu" {
					modes = modes[:1]
				}

				for _, mode := range modes {
					t.Run(mode.name, func(t *testing.T) {
						renderEnv := make([]string, len(env))
						for i, entry := range env {
							renderEnv[i] = strings.ReplaceAll(entry, "OMP_SECONDARY_TEST=init", "OMP_SECONDARY_TEST="+mode.name)
						}
						initCommand := exec.CommandContext(t.Context(), omp, append([]string{"init", sh.name, "--config=" + cfg}, mode.flags...)...)
						initCommand.Env = renderEnv
						script, err := initCommand.Output()
						require.NoError(t, err)
						path := filepath.Join(t.TempDir(), "init"+map[string]string{"pwsh": ".ps1", "nu": ".nu"}[sh.name])
						require.NoError(t, os.WriteFile(path, append(script, []byte("\n"+sh.suffix+"\n")...), 0o600))

						cmd := exec.CommandContext(t.Context(), bin, sh.args(path)...)
						for _, entry := range env {
							if strings.HasPrefix(entry, "OMP_SECONDARY_TEST=") {
								continue
							}
							cmd.Env = append(cmd.Env, entry)
						}
						cmd.Env = append(cmd.Env, "OMP_SECONDARY_TEST=source")
						got, err := cmd.CombinedOutput()
						require.NoErrorf(t, err, "%s", got)
						require.Equal(t, strings.ReplaceAll(string(want), "init", mode.name), string(got))
					})
				}
			})
		}
	}
}
