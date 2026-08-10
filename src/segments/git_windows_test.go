//go:build windows

package segments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

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

func TestEnabledInWorktreeWindows(t *testing.T) {
	cases := []struct {
		Case               string
		Pointer            string
		ExpectedProbedDir  string
		ExpectedIsWorkTree bool
	}{
		{
			Case:               "drive absolute with native separators",
			Pointer:            `C:\repo\.git\worktrees\feat`,
			ExpectedProbedDir:  `C:\repo\.git\worktrees\feat`,
			ExpectedIsWorkTree: true,
		},
		{
			Case:               "disk-relative pointer gains the volume",
			Pointer:            "/repo/.git/worktrees/feat",
			ExpectedProbedDir:  "C:/repo/.git/worktrees/feat",
			ExpectedIsWorkTree: true,
		},
		{
			Case:               "relative pointer resolves against the .git file's folder",
			Pointer:            "./.bare",
			ExpectedProbedDir:  "C:/repo/feat/.bare",
			ExpectedIsWorkTree: false,
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         `C:/repo/feat/.git`,
			ParentFolder: `C:/repo/feat`,
		}

		env := new(mock.Environment)
		env.On("FileContent", fileInfo.Path).Return(fmt.Sprintf("gitdir: %s", tc.Pointer))
		env.On("FileContent", filepath.Join(tc.ExpectedProbedDir, "gitdir")).Return(`C:/repo/feat/.git` + "\n")
		env.On("HasFilesInDir", tc.ExpectedProbedDir, "gitdir").Return(true)
		env.On("HasFilesInDir", tc.ExpectedProbedDir, "HEAD").Return(true)
		env.On("PathSeparator").Return(string(os.PathSeparator))

		g := &Git{}
		g.Init(options.Map{}, env)

		assert.True(t, g.hasWorktree(fileInfo), tc.Case)
		assert.Equal(t, tc.ExpectedIsWorkTree, g.IsWorkTree, tc.Case)
		assert.True(t, filepath.IsAbs(g.mainSCMDir), tc.Case+": mainSCMDir must be absolute")
		assert.True(t, filepath.IsAbs(g.scmDir), tc.Case+": scmDir must be absolute")
	}
}
