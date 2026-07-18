package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestVIMode(t *testing.T) {
	cases := []struct {
		Case            string
		Env             string
		ExpectedMode    string
		ExpectedKeymap  string
		ExpectedEnabled bool
	}{
		{Case: "main keymap maps to insert", Env: "main", ExpectedMode: "insert", ExpectedKeymap: "main", ExpectedEnabled: true},
		{Case: "viins keymap maps to insert", Env: "viins", ExpectedMode: "insert", ExpectedKeymap: "viins", ExpectedEnabled: true},
		{Case: "emacs keymap maps to insert", Env: "emacs", ExpectedMode: "insert", ExpectedKeymap: "emacs", ExpectedEnabled: true},
		{Case: "insert keymap maps is preserved", Env: "insert", ExpectedMode: "insert", ExpectedKeymap: "insert", ExpectedEnabled: true},
		{Case: "vicmd keymap maps to normal", Env: "vicmd", ExpectedMode: "normal", ExpectedKeymap: "vicmd", ExpectedEnabled: true},
		{Case: "default keymap maps to normal", Env: "default", ExpectedMode: "normal", ExpectedKeymap: "default", ExpectedEnabled: true},
		{Case: "visual keymap is preserved", Env: "visual", ExpectedMode: "visual", ExpectedKeymap: "visual", ExpectedEnabled: true},
		{Case: "viopp keymap is preserved", Env: "viopp", ExpectedMode: "viopp", ExpectedKeymap: "viopp", ExpectedEnabled: true},
		{Case: "operator keymap maps to viopp", Env: "operator", ExpectedMode: "viopp", ExpectedKeymap: "operator", ExpectedEnabled: true},
		{Case: "f keymap maps to viopp", Env: "f", ExpectedMode: "viopp", ExpectedKeymap: "f", ExpectedEnabled: true},
		{Case: "F keymap maps to viopp", Env: "F", ExpectedMode: "viopp", ExpectedKeymap: "F", ExpectedEnabled: true},
		{Case: "t keymap maps to viopp", Env: "t", ExpectedMode: "viopp", ExpectedKeymap: "t", ExpectedEnabled: true},
		{Case: "T keymap maps to viopp", Env: "T", ExpectedMode: "viopp", ExpectedKeymap: "T", ExpectedEnabled: true},
		{Case: "replace keymap is preserved", Env: "replace", ExpectedMode: "replace", ExpectedKeymap: "replace", ExpectedEnabled: true},
		{Case: "replace_one keymap maps to replace", Env: "replace_one", ExpectedMode: "replace", ExpectedKeymap: "replace_one", ExpectedEnabled: true},
		{Case: "unknown keymap falls through", Env: "custom", ExpectedMode: "custom", ExpectedKeymap: "custom", ExpectedEnabled: true},
		{Case: "empty env disables segment", Env: "", ExpectedEnabled: false},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", poshVIModeEnv).Return(tc.Env)

		v := &VIMode{}
		v.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, v.Enabled(), tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		assert.Equal(t, tc.ExpectedKeymap, v.Keymap, tc.Case)
		assert.Equal(t, tc.ExpectedMode, renderTemplate(env, v.Template(), v), tc.Case)
	}
}
