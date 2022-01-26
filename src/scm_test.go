package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScmStatusChanged(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Status   ScmStatus
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

func TestScmStatusUnmerged(t *testing.T) {
	expected := "x1"
	status := &ScmStatus{
		Unmerged: 1,
	}
	assert.Equal(t, expected, status.String())
}

func TestScmStatusUnmergedModified(t *testing.T) {
	expected := "~3 x1"
	status := &ScmStatus{
		Unmerged: 1,
		Modified: 3,
	}
	assert.Equal(t, expected, status.String())
}

func TestScmStatusEmpty(t *testing.T) {
	expected := ""
	status := &ScmStatus{}
	assert.Equal(t, expected, status.String())
}

func TestTruncateBranch(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   string
		Branch     string
		FullBranch bool
		MaxLength  interface{}
	}{
		{Case: "No limit", Expected: "are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: false},
		{Case: "No limit - larger", Expected: "are-belong", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, MaxLength: 10.0},
		{Case: "No limit - smaller", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 13.0},
		{Case: "Invalid setting", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: "burp"},
		{Case: "Lower than limit", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 20.0},

		{Case: "No limit - full branch", Expected: "/all-your-base/are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: true},
		{Case: "No limit - larger - full branch", Expected: "/all-your-base", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, MaxLength: 14.0},
		{Case: "No limit - smaller - full branch ", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 14.0},
		{Case: "Invalid setting - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: "burp"},
		{Case: "Lower than limit - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 20.0},
	}

	for _, tc := range cases {
		props := properties.Map{
			BranchMaxLength: tc.MaxLength,
			FullBranchPath:  tc.FullBranch,
		}
		p := &Plastic{
			scm: scm{
				props: props,
			},
		}
		assert.Equal(t, tc.Expected, p.truncateBranch(tc.Branch), tc.Case)
	}
}

func TestTruncateBranchWithSymbol(t *testing.T) {
	cases := []struct {
		Case           string
		Expected       string
		Branch         string
		FullBranch     bool
		MaxLength      interface{}
		TruncateSymbol interface{}
	}{
		{Case: "No limit", Expected: "are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, TruncateSymbol: "..."},
		{Case: "No limit - larger", Expected: "are-belong...", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, MaxLength: 10.0, TruncateSymbol: "..."},
		{Case: "No limit - smaller", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 13.0, TruncateSymbol: "..."},
		{Case: "Invalid setting", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: "burp", TruncateSymbol: "..."},
		{Case: "Lower than limit", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 20.0, TruncateSymbol: "..."},

		{Case: "No limit - full branch", Expected: "/all-your-base/are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, TruncateSymbol: "..."},
		{Case: "No limit - larger - full branch", Expected: "/all-your-base...", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, MaxLength: 14.0, TruncateSymbol: "..."},
		{Case: "No limit - smaller - full branch ", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 14.0, TruncateSymbol: "..."},
		{Case: "Invalid setting - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: "burp", TruncateSymbol: "..."},
		{Case: "Lower than limit - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 20.0, TruncateSymbol: "..."},
	}

	for _, tc := range cases {
		props := properties.Map{
			BranchMaxLength: tc.MaxLength,
			TruncateSymbol:  tc.TruncateSymbol,
			FullBranchPath:  tc.FullBranch,
		}
		p := &Plastic{
			scm: scm{
				props: props,
			},
		}
		assert.Equal(t, tc.Expected, p.truncateBranch(tc.Branch), tc.Case)
	}
}

func TestScmShouldIgnoreRootRepository(t *testing.T) {
	cases := []struct {
		Case     string
		Dir      string
		Expected bool
	}{
		{Case: "inside excluded", Dir: "/home/bill/repo"},
		{Case: "oustide excluded", Dir: "/home/melinda"},
		{Case: "excluded exact match", Dir: "/home/gates", Expected: true},
		{Case: "excluded inside match", Dir: "/home/gates/bill", Expected: true},
	}

	for _, tc := range cases {
		excludeFolders := []string{
			"/home/bill",
			"/home/gates.*",
		}
		props := properties.Map{
			properties.ExcludeFolders: excludeFolders,
		}
		env := new(mock.MockedEnvironment)
		env.On("Home").Return("/home/bill")
		env.On("GOOS").Return(environment.WindowsPlatform)
		s := &scm{
			props: props,
			env:   env,
		}
		got := s.shouldIgnoreRootRepository(tc.Dir)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
