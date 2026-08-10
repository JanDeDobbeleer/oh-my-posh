//go:build windows

package segments

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestRootPath = "C:/"

func TestResolveGitPath(t *testing.T) {
	cases := []struct {
		Case     string
		Base     string
		Path     string
		Expected string
	}{
		{
			Case:     "relative path",
			Base:     "dir\\",
			Path:     "sub",
			Expected: "dir/sub",
		},
		{
			Case:     "absolute path",
			Base:     "C:\\base",
			Path:     "C:/absolute/path",
			Expected: "C:/absolute/path",
		},
		{
			Case:     "disk-relative path",
			Base:     "C:\\base",
			Path:     "/absolute/path",
			Expected: "C:/absolute/path",
		},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, resolveGitPath(tc.Base, tc.Path), tc.Case)
	}
}

func TestWorktreeAdminIndexWindows(t *testing.T) {
	cases := []struct {
		Case     string
		Path     string
		Expected string
	}{
		{Case: "drive absolute, native separators", Path: `C:\repo\.git\worktrees\feat`, Expected: "C:/repo/.git"},
		{Case: "drive absolute, trailing separator", Path: `C:\repo\.git\worktrees\feat\`, Expected: "C:/repo/.git"},
		{Case: "drive absolute, dot component", Path: `C:\repo\.git\worktrees\feat\.`, Expected: "C:/repo/.git"},
		{Case: "UNC path", Path: `\\server\share\.git\worktrees\feat`, Expected: "//server/share/.git"},
		{Case: "two components after worktrees", Path: `C:\repo\.git\worktrees\a\b`, Expected: ""},
		{Case: "repo under a worktrees component", Path: `C:\me\worktrees\proj\.bare`, Expected: ""},
	}

	for _, tc := range cases {
		var got string
		if index := worktreeAdminIndex(tc.Path); index > -1 {
			got = filepath.ToSlash(filepath.Clean(tc.Path))[:index]
		}

		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
