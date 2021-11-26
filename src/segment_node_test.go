package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestNodeMatchesVersionFile(t *testing.T) {
	nodeVersion := version{
		Full:  "1.2.3",
		Major: "1",
		Minor: "2",
		Patch: "3",
	}
	cases := []struct {
		Case      string
		Expected  bool
		RCVersion string
	}{
		{Case: "no file context", Expected: true, RCVersion: ""},
		{Case: "version match", Expected: true, RCVersion: "1.2.3"},
		{Case: "version mismatch", Expected: false, RCVersion: "3.2.1"},
		{Case: "version match in other format", Expected: true, RCVersion: "v1.2.3"},
		{Case: "version match without patch", Expected: true, RCVersion: "1.2"},
		{Case: "version match without patch in other format", Expected: true, RCVersion: "v1.2"},
		{Case: "version match without minor", Expected: true, RCVersion: "1"},
		{Case: "version match without minor in other format", Expected: true, RCVersion: "v1"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getFileContent", ".nvmrc").Return(tc.RCVersion)
		node := &node{
			language: &language{
				env: env,
				activeCommand: &cmd{
					version: &nodeVersion,
				},
			},
		}
		assert.Equal(t, tc.Expected, node.matchesVersionFile(), tc.Case)
	}
}

func TestNodeInContext(t *testing.T) {
	cases := []struct {
		Case           string
		HasYarn        bool
		hasNPM         bool
		hasDefault     bool
		PkgMgrEnabled  bool
		ExpectedString string
	}{
		{Case: "no package manager file", ExpectedString: "", PkgMgrEnabled: true},
		{Case: "yarn", HasYarn: true, ExpectedString: "yarn", PkgMgrEnabled: true},
		{Case: "npm", hasNPM: true, ExpectedString: "npm", PkgMgrEnabled: true},
		{Case: "default", hasDefault: true, ExpectedString: "npm", PkgMgrEnabled: true},
		{Case: "disabled", HasYarn: true, ExpectedString: "", PkgMgrEnabled: false},
		{Case: "yarn and npm", HasYarn: true, hasNPM: true, ExpectedString: "yarn", PkgMgrEnabled: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasFiles", "yarn.lock").Return(tc.HasYarn)
		env.On("hasFiles", "package-lock.json").Return(tc.hasNPM)
		env.On("hasFiles", "package.json").Return(tc.hasDefault)
		node := &node{
			language: &language{
				env: env,
				props: map[Property]interface{}{
					YarnIcon:              "yarn",
					NPMIcon:               "npm",
					DisplayPackageManager: tc.PkgMgrEnabled,
				},
			},
		}
		node.loadContext()
		assert.Equal(t, tc.ExpectedString, node.packageManagerIcon, tc.Case)
	}
}
