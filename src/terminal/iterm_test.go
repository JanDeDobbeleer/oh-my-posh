package terminal

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestRenderItermFeatures(t *testing.T) {
	cases := []struct {
		Case     string
		Shell    string
		Pwd      string
		User     string
		Host     string
		Expected string
		Features ITermFeatures
	}{
		{
			Case:     "CurrentDir clean pwd",
			Features: ITermFeatures{CurrentDir},
			Shell:    shell.GENERIC,
			Pwd:      "/home/user/project",
			Expected: "\x1b]1337;CurrentDir=/home/user/project\x07",
		},
		{
			Case:     "CurrentDir malicious pwd is sanitized",
			Features: ITermFeatures{CurrentDir},
			Shell:    shell.GENERIC,
			Pwd:      maliciousPwd,
			Expected: "\x1b]1337;CurrentDir=evil\\]0;PWNEDrest\x07",
		},
		{
			Case:     "RemoteHost clean user and host",
			Features: ITermFeatures{RemoteHost},
			Shell:    shell.GENERIC,
			User:     "jan",
			Host:     "box",
			Expected: "\x1b]1337;RemoteHost=jan@box\x07",
		},
		{
			Case:     "RemoteHost malicious user and host are sanitized",
			Features: ITermFeatures{RemoteHost},
			Shell:    shell.GENERIC,
			User:     maliciousPwd,
			Host:     maliciousPwd,
			Expected: "\x1b]1337;RemoteHost=evil\\]0;PWNEDrest@evil\\]0;PWNEDrest\x07",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			Init(tc.Shell)

			got := RenderItermFeatures(tc.Features, tc.Shell, tc.Pwd, tc.User, tc.Host)

			assert.Equal(t, tc.Expected, got, tc.Case)
			assert.NotContains(t, got, "\x1b]0;PWNED\x07", tc.Case)
		})
	}
}
