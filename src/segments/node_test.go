package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/alecthomas/assert"
)

func TestNodeMatchesVersionFile(t *testing.T) {
	nodeVersion := version{
		Full:  "20.14.0",
		Major: "20",
		Minor: "14",
		Patch: "0",
	}
	cases := []struct {
		Case            string
		ExpectedVersion string
		RCVersion       string
		Expected        bool
	}{
		{Case: "no file context", Expected: true, RCVersion: ""},
		{Case: "version match", Expected: true, ExpectedVersion: "20.14.0", RCVersion: "20.14.0"},
		{Case: "version match with newline", Expected: true, ExpectedVersion: "20.14.0", RCVersion: "20.14.0\n"},
		{Case: "version mismatch", Expected: false, ExpectedVersion: "3.2.1", RCVersion: "3.2.1"},
		{Case: "version match in other format", Expected: true, ExpectedVersion: "20.14.0", RCVersion: "v20.14.0"},
		{Case: "version match without patch", Expected: true, ExpectedVersion: "20.14", RCVersion: "20.14"},
		{Case: "version match without patch in other format", Expected: true, ExpectedVersion: "20.14", RCVersion: "v20.14"},
		{Case: "version match without minor", Expected: true, ExpectedVersion: "20", RCVersion: "20"},
		{Case: "version match without minor in other format", Expected: true, ExpectedVersion: "20", RCVersion: "v20"},
		{Case: "lts match", Expected: true, ExpectedVersion: "20.14.0", RCVersion: "lts/iron"},
		{Case: "lts match upper case", Expected: true, ExpectedVersion: "20.14.0", RCVersion: "lts/Iron"},
		{Case: "lts mismatch", Expected: false, ExpectedVersion: "8.17.0", RCVersion: "lts/carbon"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", ".nvmrc").Return(tc.RCVersion)

		node := &Node{
			language: language{
				version: nodeVersion,
			},
		}
		node.Init(properties.Map{}, env)

		version, match := node.matchesVersionFile()
		assert.Equal(t, tc.Expected, match, tc.Case)
		assert.Equal(t, tc.ExpectedVersion, version, tc.Case)
	}
}

func TestNodeInContext(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		hasPNPM        bool
		hasYarn        bool
		hasNPM         bool
		hasDefault     bool
		hasBun         bool
		PkgMgrEnabled  bool
	}{
		{Case: "no package manager file", ExpectedString: "", PkgMgrEnabled: true},
		{Case: "pnpm", hasPNPM: true, ExpectedString: "pnpm", PkgMgrEnabled: true},
		{Case: "yarn", hasYarn: true, ExpectedString: "yarn", PkgMgrEnabled: true},
		{Case: "npm", hasNPM: true, ExpectedString: "npm", PkgMgrEnabled: true},
		{Case: "default", hasDefault: true, ExpectedString: "npm", PkgMgrEnabled: true},
		{Case: "disabled by pnpm", hasPNPM: true, ExpectedString: "", PkgMgrEnabled: false},
		{Case: "disabled by yarn", hasYarn: true, ExpectedString: "", PkgMgrEnabled: false},
		{Case: "pnpm and npm", hasPNPM: true, hasNPM: true, ExpectedString: "pnpm", PkgMgrEnabled: true},
		{Case: "yarn and npm", hasYarn: true, hasNPM: true, ExpectedString: "yarn", PkgMgrEnabled: true},
		{Case: "pnpm, yarn, and npm", hasPNPM: true, hasYarn: true, hasNPM: true, ExpectedString: "pnpm", PkgMgrEnabled: true},
		{Case: "bun", hasBun: true, ExpectedString: "bun", PkgMgrEnabled: true},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasFiles", "pnpm-lock.yaml").Return(tc.hasPNPM)
		env.On("HasFiles", "yarn.lock").Return(tc.hasYarn)
		env.On("HasFiles", "package-lock.json").Return(tc.hasNPM)
		env.On("HasFiles", "package.json").Return(tc.hasDefault)
		env.On("HasFiles", "bun.lockb").Return(tc.hasBun)

		props := properties.Map{
			PnpmIcon:            "pnpm",
			YarnIcon:            "yarn",
			NPMIcon:             "npm",
			BunIcon:             "bun",
			FetchPackageManager: tc.PkgMgrEnabled,
		}

		node := &Node{}
		node.Init(props, env)

		node.loadContext()
		assert.Equal(t, tc.ExpectedString, node.PackageManagerIcon, tc.Case)
	}
}
