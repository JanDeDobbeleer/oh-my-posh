package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestScmStatusChanged(t *testing.T) {
	cases := []struct {
		Case     string
		Status   ScmStatus
		Expected bool
	}{
		{
			Case:     "No changes",
			Expected: false,
			Status:   ScmStatus{},
		},
		{
			Case:     "Added",
			Expected: true,
			Status: ScmStatus{
				Added: 1,
			},
		},
		{
			Case:     "Moved",
			Expected: true,
			Status: ScmStatus{
				Moved: 1,
			},
		},
		{
			Case:     "Modified",
			Expected: true,
			Status: ScmStatus{
				Modified: 1,
			},
		},
		{
			Case:     "Deleted",
			Expected: true,
			Status: ScmStatus{
				Deleted: 1,
			},
		},
		{
			Case:     "Unmerged",
			Expected: true,
			Status: ScmStatus{
				Unmerged: 1,
			},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Status.Changed(), tc.Case)
	}
}

func TestScmStatusString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Status   ScmStatus
	}{
		{
			Case:     "Unmerged",
			Expected: "x1",
			Status: ScmStatus{
				Unmerged: 1,
			},
		},
		{
			Case:     "Unmerged and Modified",
			Expected: "~3 x1",
			Status: ScmStatus{
				Unmerged: 1,
				Modified: 3,
			},
		},
		{
			Case:   "Empty",
			Status: ScmStatus{},
		},
		{
			Case:     "Format override",
			Expected: "Added: 1",
			Status: ScmStatus{
				Added: 1,
				Formats: map[string]string{
					"Added": "Added: %d",
				},
			},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Status.String().String(), tc.Case)
	}
}

func TestHasCommand(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedCommand string
		Command         string
		GOOS            string
		IsWslSharedPath bool
		NativeFallback  bool
	}{
		{Case: "On Windows", ExpectedCommand: "git.exe", GOOS: runtime.WINDOWS},
		{Case: "Cache", ExpectedCommand: "git.exe", Command: "git.exe"},
		{Case: "Non Windows", ExpectedCommand: "git"},
		{Case: "Iside WSL2, non shared", ExpectedCommand: "git"},
		{Case: "Iside WSL2, shared", ExpectedCommand: "git.exe", IsWslSharedPath: true},
		{Case: "Iside WSL2, shared fallback", ExpectedCommand: "git", IsWslSharedPath: true, NativeFallback: true},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(tc.GOOS)
		env.On("InWSLSharedDrive").Return(tc.IsWslSharedPath)
		env.On("HasCommand", "git").Return(true)
		env.On("HasCommand", "git.exe").Return(!tc.NativeFallback)

		props := options.Map{
			NativeFallback: tc.NativeFallback,
		}

		s := &Scm{
			command: tc.Command,
		}
		s.Init(props, env)

		_ = s.hasCommand(GITCOMMAND)
		assert.Equal(t, tc.ExpectedCommand, s.command, tc.Case)
	}
}

func TestFormatBranch(t *testing.T) {
	cases := []struct {
		MappedBranches map[string]string
		Case           string
		Expected       string
		Input          string
		BranchTemplate string
		Upstream       string
	}{
		{
			Case:     "No settings",
			Input:    "main",
			Expected: "main",
		},
		{
			Case:           "BranchMaxLength higher than branch name",
			Input:          "main",
			Expected:       "main",
			BranchTemplate: "{{ trunc 5 .Branch }}",
		},
		{
			Case:           "BranchMaxLength lower than branch name",
			Input:          "feature/test-this-branch",
			Expected:       "featu",
			BranchTemplate: "{{ trunc 5 .Branch }}",
		},
		{
			Case:           "BranchMaxLength lower than branch name, with truncate symbol",
			Input:          "feature/test-this-branch",
			Expected:       "feat…",
			BranchTemplate: "{{ truncE 5 .Branch }}",
		},
		{
			Case:           "BranchMaxLength lower than branch name, with truncate symbol and no FullBranchPath",
			Input:          "feature/test-this-branch",
			Expected:       "test…",
			BranchTemplate: "{{ truncE 5 (base .Branch) }}",
		},
		{
			Case:           "BranchMaxLength lower to branch name, with truncate symbol",
			Input:          "feat",
			Expected:       "feat",
			BranchTemplate: "{{ trunc 5 .Branch }}",
		},
		{
			Case:     "Branch mapping, no BranchMaxLength",
			Input:    "feat/my-new-feature",
			Expected: "🚀 my-new-feature",
			MappedBranches: map[string]string{
				"feat/*": "🚀 ",
				"bug/*":  "🐛 ",
			},
		},
		{
			Case:           "Branch mapping, with BranchMaxLength",
			Input:          "feat/my-new-feature",
			Expected:       "🚀 my-",
			BranchTemplate: "{{ trunc 5 .Branch }}",
			MappedBranches: map[string]string{
				"feat/*": "🚀 ",
				"bug/*":  "🐛 ",
			},
		},
		{
			Case:           "Branch with upstream",
			Input:          "feat/my-new-feature",
			Expected:       "feat/my-new-feature@origin",
			Upstream:       "origin",
			BranchTemplate: "{{ .Branch }}{{ if .Upstream }}@{{ .Upstream }}{{ end }}",
		},
		{
			// a branch name is repo-controlled: its chevrons must be escaped so
			// the writer cannot parse them as anchors (the writer renders
			// <<>red<>> as literal <red>)
			Case:     "Branch name with anchor-shaped text is escaped",
			Input:    "<red>pwned-branch",
			Expected: "<<>red<>>pwned-branch",
		},
		{
			Case:     "Branch name with hyperlink markup is escaped",
			Input:    "<LINK>https://attacker.example<TEXT>x",
			Expected: "<<>LINK<>>https://attacker.example<<>TEXT<>>x",
		},
		{
			// the mapped value is user configuration and keeps its anchors
			Case:     "Mapped branch value keeps markup",
			Input:    "feat/x",
			Expected: "<#ff0000>feat</> x",
			MappedBranches: map[string]string{
				"feat/*": "<#ff0000>feat</> ",
			},
		},
		{
			// data flows through branch_template escaped; config text keeps anchors
			Case:           "Branch template escapes data, keeps config anchors",
			Input:          "<red>x",
			Expected:       "<b><<>red<>>x</>",
			BranchTemplate: "<b>{{ .Branch }}</>",
		},
	}

	for _, tc := range cases {
		props := options.Map{
			MappedBranches: tc.MappedBranches,
			BranchTemplate: tc.BranchTemplate,
		}

		s := &Scm{
			Upstream: tc.Upstream,
		}
		s.Init(props, nil)

		env := new(mock.Environment)
		env.On("Shell").Return(shell.BASH)
		template.Cache = new(cache.Template)
		template.Init(env, nil, nil)

		got := s.formatBranch(tc.Input)
		assert.Equal(t, tc.Expected, got.String(), tc.Case)
	}
}
