package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(mock.Environment)
	env.On("Shell").Return(expected, nil)
	env.On("Flags").Return(&runtime.Flags{ShellVersion: "1.2.3"})

	s := &Shell{}
	s.Init(options.Map{}, env)

	_ = s.Enabled()
	assert.Equal(t, expected, renderTemplate(env, s.Template(), s))
}

func TestUseMappedShellNames(t *testing.T) {
	cases := []struct {
		Shell    string
		Expected string
	}{
		{Shell: "zsh", Expected: "zsh"},
		{Shell: "pwsh", Expected: "PS"},
		{Shell: "PWSH", Expected: "PS"},
		{Shell: "nu", Expected: "<p:green>></>"},
		// an unmapped shell name is not user configuration: its chevrons must
		// be escaped so the writer cannot parse them as anchors.
		{Shell: "<red>evil</>", Expected: "<<>red<>>evil<<>/<>>"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return(tc.Shell, nil)
		env.On("Flags").Return(&runtime.Flags{ShellVersion: "1.2.3"})

		props := options.Map{
			MappedShellNames: map[string]string{"pwsh": "PS", "nu": "<p:green>></>"},
		}

		s := &Shell{}
		s.Init(props, env)

		_ = s.Enabled()
		got := renderTemplate(env, s.Template(), s)
		assert.Equal(t, tc.Expected, got, tc.Shell)
	}
}
