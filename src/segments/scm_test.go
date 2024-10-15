package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

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
		assert.Equal(t, tc.Expected, tc.Status.String(), tc.Case)
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
		s := &scm{
			env: env,
			props: properties.Map{
				NativeFallback: tc.NativeFallback,
			},
			command: tc.Command,
		}

		_ = s.hasCommand(GITCOMMAND)
		assert.Equal(t, tc.ExpectedCommand, s.command, tc.Case)
	}
}

func TestFormatBranch(t *testing.T) {
	cases := []struct {
		MappedBranches   map[string]string
		Case             string
		Expected         string
		Input            string
		TruncateSymbol   string
		BranchMaxLength  int
		NoFullBranchPath bool
	}{
		{
			Case:     "No settings",
			Input:    "main",
			Expected: "main",
		},
		{
			Case:            "BranchMaxLength higher than branch name",
			Input:           "main",
			Expected:        "main",
			BranchMaxLength: 10,
		},
		{
			Case:            "BranchMaxLength lower than branch name",
			Input:           "feature/test-this-branch",
			Expected:        "featu",
			BranchMaxLength: 5,
		},
		{
			Case:            "BranchMaxLength lower than branch name, with truncate symbol",
			Input:           "feature/test-this-branch",
			Expected:        "feat…",
			BranchMaxLength: 5,
			TruncateSymbol:  "…",
		},
		{
			Case:             "BranchMaxLength lower than branch name, with truncate symbol and no FullBranchPath",
			Input:            "feature/test-this-branch",
			Expected:         "test…",
			BranchMaxLength:  5,
			TruncateSymbol:   "…",
			NoFullBranchPath: true,
		},
		{
			Case:            "BranchMaxLength lower to branch name, with truncate symbol",
			Input:           "feat",
			Expected:        "feat",
			BranchMaxLength: 5,
			TruncateSymbol:  "…",
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
			Case:            "Branch mapping, with BranchMaxLength",
			Input:           "feat/my-new-feature",
			Expected:        "🚀 my-",
			BranchMaxLength: 5,
			MappedBranches: map[string]string{
				"feat/*": "🚀 ",
				"bug/*":  "🐛 ",
			},
		},
	}

	for _, tc := range cases {
		g := &Git{
			scm: scm{
				props: properties.Map{
					MappedBranches:  tc.MappedBranches,
					BranchMaxLength: tc.BranchMaxLength,
					TruncateSymbol:  tc.TruncateSymbol,
					FullBranchPath:  !tc.NoFullBranchPath,
				},
			},
		}

		got := g.formatBranch(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
