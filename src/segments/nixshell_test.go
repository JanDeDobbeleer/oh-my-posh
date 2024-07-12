package segments

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
)

const (
	nixPath     = "/nix/store/zznw8fnzss1vaqfg5hmv3y79s3hkqczi-devshell-dir/bin"
	defaultPath = "/users/xyz/testing"
	fullNixPath = defaultPath + ":" + nixPath
)

func TestNixShellSegment(t *testing.T) {
	cases := []struct {
		name           string
		expectedString string
		shellType      string
		enabled        bool
	}{
		{
			name:           "Pure Nix Shell",
			expectedString: "via pure-shell",
			shellType:      "pure",
			enabled:        true,
		},
		{
			name:           "Impure Nix Shell",
			expectedString: "via impure-shell",
			shellType:      "impure",
			enabled:        true,
		},
		{
			name:           "Unknown Nix Shell",
			expectedString: "via unknown-shell",
			shellType:      "unknown",
			enabled:        true,
		},
		{
			name:           "No Nix Shell",
			expectedString: "",
			shellType:      "",
			enabled:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("Getenv", "IN_NIX_SHELL").Return(tc.shellType)

			path := defaultPath
			if tc.shellType != "" {
				path = fullNixPath
			}

			env.On("Getenv", "PATH").Return(path)

			n := NixShell{}
			n.Init(properties.Map{}, env)

			assert.Equal(t, tc.enabled, n.Enabled(), fmt.Sprintf("Failed in case: %s", tc.name))

			if tc.enabled {
				assert.Equal(t, tc.expectedString, renderTemplate(env, n.Template(), n), tc.name)
			}
		})
	}
}
