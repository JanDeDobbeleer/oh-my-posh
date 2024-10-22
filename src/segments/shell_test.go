package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(mock.Environment)
	env.On("Shell").Return(expected, nil)
	env.On("Flags").Return(&runtime.Flags{ShellVersion: "1.2.3"})

	s := &Shell{}
	s.Init(properties.Map{}, env)

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
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return(tc.Expected, nil)
		env.On("Flags").Return(&runtime.Flags{ShellVersion: "1.2.3"})

		props := properties.Map{
			MappedShellNames: map[string]string{"pwsh": "PS"},
		}

		s := &Shell{}
		s.Init(props, env)

		_ = s.Enabled()
		got := renderTemplate(env, s.Template(), s)
		assert.Equal(t, tc.Expected, got)
	}
}
