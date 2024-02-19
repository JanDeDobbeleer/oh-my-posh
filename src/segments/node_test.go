package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/alecthomas/assert"
)

func TestNodeMatchesVersionFile(t *testing.T) {
	nodeVersion := version{
		Full:  "20.11.1",
		Major: "20",
		Minor: "11",
		Patch: "1",
	}
	cases := []struct {
		Case            string
		Expected        bool
		ExpectedVersion string
		RCVersion       string
	}{
		{Case: "no file context", Expected: true, RCVersion: ""},
		{Case: "version match", Expected: true, ExpectedVersion: "20.11.1", RCVersion: "20.11.1"},
		{Case: "version match with newline", Expected: true, ExpectedVersion: "20.11.1", RCVersion: "20.11.1\n"},
		{Case: "version mismatch", Expected: false, ExpectedVersion: "3.2.1", RCVersion: "3.2.1"},
		{Case: "version match in other format", Expected: true, ExpectedVersion: "20.11.1", RCVersion: "v20.11.1"},
		{Case: "version match without patch", Expected: true, ExpectedVersion: "20.11", RCVersion: "20.11"},
		{Case: "version match without patch in other format", Expected: true, ExpectedVersion: "20.11", RCVersion: "v20.11"},
		{Case: "version match without minor", Expected: true, ExpectedVersion: "20", RCVersion: "20"},
		{Case: "version match without minor in other format", Expected: true, ExpectedVersion: "20", RCVersion: "v20"},
		{Case: "lts match", Expected: true, ExpectedVersion: "20.11.1", RCVersion: "lts/iron"},
		{Case: "lts match upper case", Expected: true, ExpectedVersion: "20.11.1", RCVersion: "lts/Iron"},
		{Case: "lts mismatch", Expected: false, ExpectedVersion: "8.17.0", RCVersion: "lts/carbon"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("FileContent", ".nvmrc").Return(tc.RCVersion)

		node := &Node{
			language: language{
				env:     env,
				version: nodeVersion,
			},
		}
		version, match := node.matchesVersionFile()
		assert.Equal(t, tc.Expected, match, tc.Case)
		assert.Equal(t, tc.ExpectedVersion, version, tc.Case)
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
		env := new(mock.MockedEnvironment)
		env.On("HasFiles", "yarn.lock").Return(tc.HasYarn)
		env.On("HasFiles", "package-lock.json").Return(tc.hasNPM)
		env.On("HasFiles", "package.json").Return(tc.hasDefault)
		node := &Node{
			language: language{
				env: env,
				props: properties.Map{
					YarnIcon:            "yarn",
					NPMIcon:             "npm",
					FetchPackageManager: tc.PkgMgrEnabled,
				},
			},
		}
		node.loadContext()
		assert.Equal(t, tc.ExpectedString, node.PackageManagerIcon, tc.Case)
	}
}
