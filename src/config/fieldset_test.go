package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolveFieldSets(t *testing.T) {
	cases := []struct {
		Config             *Config
		Case               string
		ExpectedFields     []string
		ExpectedAnalyzable bool
	}{
		{
			Case: "own template only",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
				}}},
			},
			ExpectedFields:     []string{"HEAD"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "writer default template when none is configured",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT},
				}}},
			},
			ExpectedFields:     []string{"BranchStatus", "HEAD", "Staging", "Working"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "foreground and background templates contribute",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{
						Type:                GIT,
						Template:            "{{ .HEAD }}",
						ForegroundTemplates: []string{"{{ if .Working.Changed }}p:red{{ end }}"},
						BackgroundTemplates: []string{"{{ if gt .Ahead 0 }}p:blue{{ end }}"},
					},
				}}},
			},
			ExpectedFields:     []string{"Ahead", "HEAD", "Working"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "cross-segment reference from another segment",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
					{Type: TEXT, Template: "{{ .Segments.Git.Working.Changed }}"},
				}}},
			},
			ExpectedFields:     []string{"HEAD", "Working"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "cross-segment reference resolves the alias",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Alias: "scm", Template: "{{ .HEAD }}"},
					{Type: TEXT, Template: "{{ .Segments.scm.Ahead }}"},
				}}},
			},
			ExpectedFields:     []string{"Ahead", "HEAD"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "transient prompt reference counts",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
				}}},
				TransientPrompt: &Segment{Template: "{{ .Segments.Git.Staging.Changed }}"},
			},
			ExpectedFields:     []string{"HEAD", "Staging"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "console title reference counts",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
				}}},
				ConsoleTitleTemplate: "{{ .Segments.Git.UpstreamIcon }}",
			},
			ExpectedFields:     []string{"HEAD", "UpstreamIcon"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "tooltip reference counts",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
				}}},
				Tooltips: []*Segment{
					{Type: TEXT, Tips: []string{"git"}, Template: "{{ .Segments.Git.User.Name }}"},
				},
			},
			ExpectedFields:     []string{"HEAD", "User"},
			ExpectedAnalyzable: true,
		},
		{
			Case: "whole-dot template defeats own analysis",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ . }}"},
				}}},
			},
			ExpectedFields:     []string{},
			ExpectedAnalyzable: false,
		},
		{
			Case: "whole-segment reference defeats analysis for the target",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
					{Type: TEXT, Template: "{{ .Segments.Git }}"},
				}}},
			},
			ExpectedFields:     []string{"HEAD"},
			ExpectedAnalyzable: false,
		},
		{
			Case: "opaque global template defeats every segment",
			Config: &Config{
				Blocks: []*Block{{Segments: []*Segment{
					{Type: GIT, Template: "{{ .HEAD }}"},
				}}},
				TransientPrompt: &Segment{Template: "{{ . }}"},
			},
			ExpectedFields:     []string{"HEAD"},
			ExpectedAnalyzable: false,
		},
	}

	for _, tc := range cases {
		tc.Config.ResolveFieldSets()

		segment := tc.Config.Blocks[0].Segments[0]
		assert.Equal(t, tc.ExpectedFields, segment.ReferencedFields, tc.Case)
		assert.Equal(t, tc.ExpectedAnalyzable, segment.FieldsAnalyzable, tc.Case)
	}
}

// TestGitFetchDerivedFromTemplates proves the wiring end to end: config
// analysis -> MapSegmentWithWriter -> git's Enabled(). A config whose
// templates never show a status field must not spawn the status probe, while
// one that does show it must - without either config setting fetch_status.
// This is the deliberate contract change of template-derived fetching: status
// data used to require the explicit option even when the template displayed
// its fields.
func TestGitFetchDerivedFromTemplates(t *testing.T) {
	const repoRoot = "/repo"

	statusArgs := []string{"status", "-unormal", "--branch", "--porcelain=2"}
	gitArgs := append([]string{"-C", repoRoot, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, statusArgs...)

	cases := []struct {
		Case         string
		Template     string
		ExpectStatus bool
	}{
		{
			Case:         "HEAD only skips the status probe",
			Template:     "{{ .HEAD }}",
			ExpectStatus: false,
		},
		{
			Case:         "working changes trigger the status probe",
			Template:     "{{ .HEAD }}{{ if .Working.Changed }}!{{ end }}",
			ExpectStatus: true,
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/repo/.git",
			ParentFolder: repoRoot,
			IsDir:        true,
		}

		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", "git").Return(true)
		env.On("GOOS").Return("")
		env.On("IsWsl").Return(false)
		env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
		env.On("FileContent", "/repo/.git/HEAD").Return("ref: refs/heads/main")
		env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)
		env.On("HasFolder", testify_.Anything).Return(false)
		env.On("HasFilesInDir", testify_.Anything, testify_.Anything).Return(false)
		env.MockGitCommand(repoRoot, "", statusArgs...)

		cfg := &Config{
			Blocks: []*Block{{Segments: []*Segment{
				{Type: GIT, Template: tc.Template},
			}}},
		}

		cfg.ResolveFieldSets()

		segment := cfg.Blocks[0].Segments[0]
		require.NoError(t, segment.MapSegmentWithWriter(env), tc.Case)
		assert.True(t, segment.writer.Enabled(), tc.Case)

		if tc.ExpectStatus {
			env.AssertCalled(t, "RunCommand", "git", gitArgs)
			continue
		}

		env.AssertNotCalled(t, "RunCommand", "git", gitArgs)
	}
}

// countAnalyzerCalls swaps the field-set analyzer for a counting wrapper so a
// test can assert whether ResolveFieldSets actually analyzed anything.
func countAnalyzerCalls(t *testing.T) *int {
	t.Helper()

	origAnalyzer := analyzeFields
	t.Cleanup(func() { analyzeFields = origAnalyzer })

	count := new(int)
	analyzeFields = func(text string) *template.Refs {
		*count++
		return origAnalyzer(text)
	}

	return count
}

func TestFieldSetStampsSurviveGobRoundTrip(t *testing.T) {
	cfg := &Config{
		Blocks: []*Block{{Segments: []*Segment{
			{Type: GIT, Template: "{{ .HEAD }}"},
		}}},
	}

	cfg.ResolveFieldSets()

	restored := &Config{}
	require.NoError(t, restored.Restore(cfg.Base64()))

	assert.True(t, restored.FieldSetsResolved, "the stamped marker must round-trip")

	segment := restored.Blocks[0].Segments[0]
	assert.Equal(t, []string{"HEAD"}, segment.ReferencedFields)
	assert.True(t, segment.FieldsAnalyzable)

	count := countAnalyzerCalls(t)

	restored.ResolveFieldSets()
	assert.Zero(t, *count, "a restored, stamped config must skip re-analysis")
}

func TestGetUpgradesLegacyUnstampedCacheEntry(t *testing.T) {
	resetCache(t)

	// a pre-stamp binary cached the config without field sets: same payload
	// shape, FieldSetsResolved never set
	legacy := &Config{
		Blocks: []*Block{{Segments: []*Segment{
			{Type: GIT, Template: "{{ .HEAD }}"},
		}}},
	}
	cache.Session.Set(configKey, legacy.Base64(), cache.INFINITE)

	count := countAnalyzerCalls(t)

	cfg := Get("", false)
	assert.True(t, cfg.FieldSetsResolved, "a legacy entry must be analyzed on restore")
	assert.Positive(t, *count, "the legacy entry requires one analysis")
	assert.Equal(t, []string{"HEAD"}, cfg.Blocks[0].Segments[0].ReferencedFields)

	// the upgrade re-stored the entry, so the next restore skips analysis
	*count = 0

	cfg = Get("", false)
	assert.True(t, cfg.FieldSetsResolved)
	assert.Zero(t, *count, "the upgraded entry must restore stamped")
	assert.Equal(t, []string{"HEAD"}, cfg.Blocks[0].Segments[0].ReferencedFields)
}

func TestGetReloadRefreshesFieldSets(t *testing.T) {
	resetCache(t)

	writeConfig := func(template string) string {
		t.Helper()

		file := filepath.Join(t.TempDir(), "reload.omp.json")
		data := fmt.Sprintf(`{"version":3,"blocks":[{"type":"prompt","segments":[{"type":"git","template":"%s"}]}]}`, template)
		require.NoError(t, os.WriteFile(file, []byte(data), 0o600))

		return file
	}

	file := writeConfig("{{ .HEAD }}")
	cache.Session.Set(SourceKey, file, cache.INFINITE)

	cfg := Get(file, true)
	assert.Equal(t, []string{"HEAD"}, cfg.Blocks[0].Segments[0].ReferencedFields)

	// same session, edited config content: reload must re-analyze
	file = writeConfig("{{ .HEAD }}{{ if .Working.Changed }}!{{ end }}")
	cache.Session.Set(SourceKey, file, cache.INFINITE)

	cfg = Get(file, true)
	assert.Equal(t, []string{"HEAD", "Working"}, cfg.Blocks[0].Segments[0].ReferencedFields)
	assert.True(t, cfg.FieldSetsResolved)
}
