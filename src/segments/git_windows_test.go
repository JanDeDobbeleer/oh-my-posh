//go:build windows

package segments

import (
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
