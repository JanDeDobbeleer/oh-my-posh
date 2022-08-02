//go:build !windows

package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestRootPath = "/"

func TestResolveGitPath(t *testing.T) {
	cases := []struct {
		Case     string
		Base     string
		Path     string
		Expected string
	}{
		{
			Case:     "relative path",
			Base:     "dir/",
			Path:     "sub",
			Expected: "dir/sub",
		},
		{
			Case:     "absolute path",
			Base:     "/base",
			Path:     "/absolute/path",
			Expected: "/absolute/path",
		},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.Expected, resolveGitPath(tc.Base, tc.Path), tc.Case)
	}
}
