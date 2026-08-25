package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatuslinePWD(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))

	child := filepath.Join(dir, "child")
	require.NoError(t, os.Mkdir(child, 0o755))

	// filepath.Join would clean the dot component before the test sees it.
	dotted := child + string(os.PathSeparator) + ".."

	link := filepath.Join(dir, "link")
	symlinked := true
	if err := os.Symlink(child, link); err != nil {
		symlinked = false
	}

	cases := []struct {
		Case      string
		Candidate string
		Expected  string
		Skip      bool
	}{
		{Case: "existing directory", Candidate: dir, Expected: dir},
		{Case: "empty", Candidate: "", Expected: ""},
		{Case: "relative", Candidate: "sub/deep", Expected: ""},
		{Case: "missing", Candidate: filepath.Join(dir, "nope"), Expected: ""},
		{Case: "a file", Candidate: file, Expected: ""},
		{Case: "dot components kept verbatim", Candidate: dotted, Expected: dotted},
		{Case: "symlink to a directory kept verbatim", Candidate: link, Expected: link, Skip: !symlinked},
	}

	for _, tc := range cases {
		if tc.Skip {
			continue
		}

		assert.Equal(t, tc.Expected, statuslinePWD(tc.Candidate), tc.Case)
	}
}

func TestClaudePWD(t *testing.T) {
	cases := []struct {
		Case     string
		Data     *segments.ClaudeData
		Expected string
	}{
		{
			Case:     "both set, disagreeing: current_dir wins",
			Data:     &segments.ClaudeData{CWD: "/b", Workspace: segments.ClaudeWorkspace{CurrentDir: "/a"}},
			Expected: "/a",
		},
		{
			Case:     "only current_dir",
			Data:     &segments.ClaudeData{Workspace: segments.ClaudeWorkspace{CurrentDir: "/a"}},
			Expected: "/a",
		},
		{
			Case:     "only cwd",
			Data:     &segments.ClaudeData{CWD: "/b"},
			Expected: "/b",
		},
		{
			Case:     "neither",
			Data:     &segments.ClaudeData{},
			Expected: "",
		},
		{
			Case:     "project_dir is not a fallback",
			Data:     &segments.ClaudeData{Workspace: segments.ClaudeWorkspace{ProjectDir: "/p"}},
			Expected: "",
		},
		{
			Case:     "worktree.path is not a fallback",
			Data:     &segments.ClaudeData{Worktree: &segments.ClaudeWorktree{Path: "/w"}},
			Expected: "",
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, claudePWD(tc.Data), tc.Case)
	}
}

func TestCopilotPWD(t *testing.T) {
	cases := []struct {
		Case     string
		Data     *segments.CopilotCLIData
		Expected string
	}{
		{
			Case:     "both set, disagreeing: current_dir wins",
			Data:     &segments.CopilotCLIData{CWD: "/b", Workspace: segments.CopilotCLIWorkspace{CurrentDir: "/a"}},
			Expected: "/a",
		},
		{
			Case:     "only cwd",
			Data:     &segments.CopilotCLIData{CWD: "/b"},
			Expected: "/b",
		},
		{
			Case:     "neither",
			Data:     &segments.CopilotCLIData{},
			Expected: "",
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, copilotPWD(tc.Data), tc.Case)
	}
}

func TestRunStatuslineRendersToWriter(t *testing.T) {
	t.Setenv("OMP_CACHE_DIR", t.TempDir())
	t.Setenv("POSH_SESSION_ID", "d4-writer")

	cfgPath := filepath.Join(t.TempDir(), "writer.omp.yaml")
	cfgBody := `version: 3
final_space: false
blocks:
  - type: prompt
    alignment: left
    segments:
      - type: text
        style: plain
        template: 'SEAM-OK'
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgBody), 0o600))

	payload, err := json.Marshal(segments.ClaudeData{SessionID: "d4-writer"})
	require.NoError(t, err)

	template.ResetCache()

	var out bytes.Buffer
	runStatusline(bytes.NewReader(payload), &out, cfgPath, shell.CLAUDE, cache.CLAUDECACHE,
		func(d *segments.ClaudeData) string { return d.SessionID }, claudePWD, config.Claude)

	assert.Contains(t, out.String(), "SEAM-OK")
}

func TestStatuslineFlags(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		Case     string
		Data     *segments.ClaudeData
		Expected string
	}{
		{
			Case:     "usable current_dir reaches Flags.PWD",
			Data:     &segments.ClaudeData{Workspace: segments.ClaudeWorkspace{CurrentDir: dir}},
			Expected: dir,
		},
		{
			Case:     "nil data leaves PWD empty",
			Data:     nil,
			Expected: "",
		},
		{
			Case:     "non-empty but invalid primary fails closed",
			Data:     &segments.ClaudeData{Workspace: segments.ClaudeWorkspace{CurrentDir: filepath.Join(dir, "nope")}, CWD: dir},
			Expected: "",
		},
	}

	for _, tc := range cases {
		flags := statuslineFlags("cfg.yaml", shell.CLAUDE, tc.Data, claudePWD)

		assert.Equal(t, tc.Expected, flags.PWD, tc.Case)
		assert.Equal(t, "cfg.yaml", flags.ConfigPath, tc.Case)
		assert.Equal(t, shell.CLAUDE, flags.Shell, tc.Case)

		env := &runtime.Terminal{}
		env.Init(flags)

		if tc.Expected != "" {
			assert.Equal(t, tc.Expected, env.Pwd(), tc.Case)
		}
	}
}

func TestRunStatuslineUsesPayloadPWD(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("OMP_CACHE_DIR", t.TempDir())
	t.Setenv("POSH_SESSION_ID", "d4-integration")

	root := t.TempDir()
	repo := filepath.Join(root, "payload-repo")

	runStatuslineGit(t, root, "init", "-q", "-b", "payload-branch", repo)
	runStatuslineGit(t, repo, "config", "user.email", "test@example.com")
	runStatuslineGit(t, repo, "config", "user.name", "Test")
	runStatuslineGit(t, repo, "config", "core.autocrlf", "false")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644))
	runStatuslineGit(t, repo, "add", ".")
	runStatuslineGit(t, repo, "commit", "-q", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.txt"), []byte("changed\n"), 0o644))

	cfgPath := filepath.Join(t.TempDir(), "d4.omp.yaml")
	cfgBody := `version: 3
final_space: false
blocks:
  - type: prompt
    alignment: left
    segments:
      - type: text
        style: plain
        template: 'Folder=[{{ .Folder }}]'
      - type: git
        style: plain
        properties:
          fetch_status: true
        template: ' Ref=[{{ .Ref }}] Dirty=[{{ .Working.Changed }}]'
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgBody), 0o600))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NotEqual(t, repo, cwd, "the payload directory must differ from the process directory")

	// Marshal the payload so Windows path separators are escaped correctly.
	payload, err := json.Marshal(segments.ClaudeData{
		SessionID: "d4-integration",
		CWD:       repo,
		Workspace: segments.ClaudeWorkspace{CurrentDir: repo},
	})
	require.NoError(t, err)

	template.ResetCache()

	var out bytes.Buffer
	runStatusline(bytes.NewReader(payload), &out, cfgPath, shell.CLAUDE, cache.CLAUDECACHE,
		func(d *segments.ClaudeData) string { return d.SessionID }, claudePWD, config.Claude)

	rendered := out.String()

	assert.Contains(t, rendered, "Folder=["+filepath.Base(repo)+"]", rendered)
	assert.Contains(t, rendered, "Ref=[payload-branch]", rendered)
	assert.Contains(t, rendered, "Dirty=[true]", rendered)
}

func runStatuslineGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")

	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s failed: %s", strings.Join(args, " "), out)
}
