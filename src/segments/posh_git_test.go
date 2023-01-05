package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestPoshGitSegment(t *testing.T) {
	cases := []struct {
		Case              string
		PoshGitJSON       string
		FetchUpstreamIcon bool
		Template          string
		ExpectedString    string
		ExpectedEnabled   bool
	}{
		{
			Case:            "no status",
			PoshGitJSON:     "",
			ExpectedString:  "my prompt",
			ExpectedEnabled: false,
		},
		{
			Case:            "invalid data",
			PoshGitJSON:     "{",
			ExpectedString:  "my prompt",
			ExpectedEnabled: false,
		},
		{
			Case: "Changes in Working",
			PoshGitJSON: `
			{
				"RepoName": "oh-my-posh",
				"HasIndex": false,
				"GitDir": "/Users/bill/Code/oh-my-posh/.git",
				"Upstream": "origin/posh-git-json",
				"UpstreamGone": false,
				"HasUntracked": false,
				"AheadBy": 0,
				"StashCount": 0,
				"HasWorking": true,
				"BehindBy": 0,
				"Index": {
					"value": [],
					"Added": [],
					"Modified": [],
					"Deleted": [],
					"Unmerged": []
				},
				"Working": {
					"value": [
						"../src/segments/git_test.go",
						"../src/segments/posh_git_test.go"
					],
					"Added": [],
					"Modified": [
						"../src/segments/git_test.go",
						"../src/segments/posh_git_test.go"
					],
					"Deleted": [],
					"Unmerged": []
				},
				"Branch": "posh-git-json"
			}
			`,
			ExpectedString:  "\ue0a0posh-git-json ≡ \uf044 ~2",
			ExpectedEnabled: true,
		},
		{
			Case: "Changes in Working and Staging, branch ahead an behind",
			PoshGitJSON: `
			{
				"RepoName": "oh-my-posh",
				"HasIndex": false,
				"GitDir": "/Users/bill/Code/oh-my-posh/.git",
				"Upstream": "origin/posh-git-json",
				"UpstreamGone": false,
				"HasUntracked": false,
				"AheadBy": 1,
				"StashCount": 2,
				"HasWorking": true,
				"BehindBy": 1,
				"Index": {
					"value": [
						"../src/segments/git_test.go",
						"../src/segments/posh_git_test.go"
					],
					"Added": [],
					"Deleted": [
					"../src/segments/git_test.go",
					"../src/segments/posh_git_test.go"
					],
					"Modified": [],
					"Unmerged": []
				},
				"Working": {
					"value": [
						"../src/segments/git_test.go",
						"../src/segments/posh_git_test.go"
					],
					"Added": [],
					"Modified": [
						"../src/segments/git_test.go",
						"../src/segments/posh_git_test.go"
					],
					"Deleted": [],
					"Unmerged": []
				},
				"Branch": "posh-git-json"
			}
			`,
			ExpectedString:  "\ue0a0posh-git-json ↑1 ↓1 \uf044 ~2 | \uf046 -2",
			ExpectedEnabled: true,
		},
		{
			Case: "Clean branch, no upstream and stash count",
			PoshGitJSON: `
			{
				"RepoName": "oh-my-posh",
				"GitDir": "/Users/bill/Code/oh-my-posh/.git",
				"StashCount": 2,
				"Index": {
					"value": [],
					"Added": [],
					"Modified": [],
					"Deleted": [],
					"Unmerged": []
				},
				"Working": {
					"value": [],
					"Added": [],
					"Modified": [],
					"Deleted": [],
					"Unmerged": []
				},
				"Branch": "posh-git-json"
			}
			`,
			ExpectedString:  "\ue0a0posh-git-json ≢",
			ExpectedEnabled: true,
		},
		{
			Case: "No working data",
			PoshGitJSON: `
			{
				"RepoName": "oh-my-posh",
				"GitDir": "/Users/bill/Code/oh-my-posh/.git",
				"StashCount": 2,
				"Index": {
					"value": [],
					"Added": [],
					"Modified": [],
					"Deleted": [],
					"Unmerged": []
				},
				"Branch": "posh-git-json"
			}
			`,
			ExpectedString:  "\ue0a0posh-git-json ≢",
			ExpectedEnabled: true,
		},
		{
			Case:     "Fetch upstream icon (GitHub)",
			Template: "{{ .UpstreamIcon }}",
			PoshGitJSON: `
			{
				"RepoName": "oh-my-posh",
				"GitDir": "/Users/bill/Code/oh-my-posh/.git",
				"Branch": "\ue0a0posh-git-json",
				"Upstream": "origin/posh-git-json"
			}
			`,
			ExpectedString:    "\uf408",
			FetchUpstreamIcon: true,
			ExpectedEnabled:   true,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Getenv", poshGitEnv).Return(tc.PoshGitJSON)
		env.On("Home").Return("/Users/bill")
		env.On("GOOS").Return(platform.LINUX)
		env.On("RunCommand", "git", []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "remote", "get-url", "origin"}).Return("github.com/cli", nil)
		g := &Git{
			scm: scm{
				env: env,
				props: &properties.Map{
					FetchUpstreamIcon: tc.FetchUpstreamIcon,
				},
				command: GITCOMMAND,
			},
		}
		if len(tc.Template) == 0 {
			tc.Template = g.Template()
		}
		assert.Equal(t, tc.ExpectedEnabled, g.hasPoshGitStatus(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, g), tc.Case)
		}
	}
}

func TestParsePoshGitHEAD(t *testing.T) {
	cases := []struct {
		Case           string
		HEAD           string
		ExpectedString string
	}{
		{
			Case:           "branch",
			HEAD:           "main",
			ExpectedString: "\ue0a0main",
		},
		{
			Case:           "tag",
			HEAD:           "(tag)",
			ExpectedString: "\uf412tag",
		},
		{
			Case:           "commit",
			HEAD:           "(commit...)",
			ExpectedString: "\uf417commit",
		},
	}

	for _, tc := range cases {
		g := &Git{
			scm: scm{
				props: &properties.Map{},
			},
		}
		assert.Equal(t, tc.ExpectedString, g.parsePoshGitHEAD(tc.HEAD), tc.Case)
	}
}
