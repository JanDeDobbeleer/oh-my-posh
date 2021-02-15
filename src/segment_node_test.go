package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestNodeMatchesVersionFile(t *testing.T) {
	cases := []struct {
		Case      string
		Expected  bool
		RCVersion string
		Version   string
	}{
		{Case: "no file context", Expected: true, RCVersion: "", Version: "durp"},
		{Case: "version match", Expected: true, RCVersion: "durp", Version: "durp"},
		{Case: "version mismatch", Expected: false, RCVersion: "werp", Version: "durp"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getFileContent", ".nvmrc").Return(tc.RCVersion)
		node := &node{
			language: &language{
				env: env,
				activeCommand: &cmd{
					version: &version{
						full: tc.Version,
					},
				},
			},
		}
		assert.Equal(t, tc.Expected, node.matchesVersionFile(), tc.Case)
	}
}
