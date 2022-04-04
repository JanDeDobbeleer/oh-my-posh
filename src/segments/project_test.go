package segments

import (
	"fmt"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/alecthomas/assert"

	testify_mock "github.com/stretchr/testify/mock"
)

type MockData struct {
	Name            string
	Case            string
	ExpectedString  string
	PackageContents string
	File            string
}

func getMockedPackageEnv(tc *MockData) (*mock.MockedEnvironment, properties.Map) {
	env := new(mock.MockedEnvironment)
	props := properties.Map{}
	env.On("HasFiles", testify_mock.Anything).Run(func(args testify_mock.Arguments) {
		for _, c := range env.ExpectedCalls {
			if c.Method == "HasFiles" {
				c.ReturnArguments = testify_mock.Arguments{args.Get(0).(string) == tc.File}
			}
		}
	})
	env.On("FileContent", tc.File).Return(tc.PackageContents)
	return env, props
}

func TestPackage(t *testing.T) {
	cases := []*MockData{
		{Case: "1.0.0 node.js", ExpectedString: "\uf487 1.0.0 test", Name: "node", File: "package.json", PackageContents: "{\"version\":\"1.0.0\",\"name\":\"test\"}"},
		{Case: "1.0.0 php", ExpectedString: "\uf487 1.0.0 test", Name: "php", File: "composer.json", PackageContents: "{\"version\":\"1.0.0\",\"name\":\"test\"}"},
		{Case: "3.2.1 node.js", ExpectedString: "\uf487 3.2.1 test", Name: "node", File: "package.json", PackageContents: "{\"version\":\"3.2.1\",\"name\":\"test\"}"},
		{Case: "1.0.0 cargo", ExpectedString: "\uf487 1.0.0 test", Name: "cargo", File: "Cargo.toml", PackageContents: "[package]\nname=\"test\"\nversion=\"1.0.0\"\n"},
		{Case: "3.2.1 cargo", ExpectedString: "\uf487 3.2.1 test", Name: "cargo", File: "Cargo.toml", PackageContents: "[package]\nname=\"test\"\nversion=\"3.2.1\"\n"},
		{Case: "1.0.0 poetry", ExpectedString: "\uf487 1.0.0 test", Name: "poetry", File: "pyproject.toml", PackageContents: "[tool.poetry]\nname=\"test\"\nversion=\"1.0.0\"\n"},
		{Case: "3.2.1 poetry", ExpectedString: "\uf487 3.2.1 test", Name: "poetry", File: "pyproject.toml", PackageContents: "[tool.poetry]\nname=\"test\"\nversion=\"3.2.1\"\n"},
		{Case: "No version present node.js", ExpectedString: "test", Name: "node", File: "package.json", PackageContents: "{\"name\":\"test\"}"},
		{Case: "No version present cargo", ExpectedString: "test", Name: "cargo", File: "Cargo.toml", PackageContents: "[package]\nname=\"test\"\n"},
		{Case: "No version present poetry", ExpectedString: "test", Name: "poetry", File: "pyproject.toml", PackageContents: "[tool.poetry]\nname=\"test\"\n"},
		{Case: "No name present node.js", ExpectedString: "\uf487 1.0.0", Name: "node", File: "package.json", PackageContents: "{\"version\":\"1.0.0\"}"},
		{Case: "No name present cargo", ExpectedString: "\uf487 1.0.0", Name: "cargo", File: "Cargo.toml", PackageContents: "[package]\nversion=\"1.0.0\"\n"},
		{Case: "No name present poetry", ExpectedString: "\uf487 1.0.0", Name: "poetry", File: "pyproject.toml", PackageContents: "[tool.poetry]\nversion=\"1.0.0\"\n"},
		{Case: "Empty project package node.js", ExpectedString: "", Name: "node", File: "package.json", PackageContents: "{}"},
		{Case: "Empty project package cargo", ExpectedString: "", Name: "cargo", File: "Cargo.toml", PackageContents: ""},
		{Case: "Empty project package poetry", ExpectedString: "", Name: "poetry", File: "pyproject.toml", PackageContents: ""},
		{Case: "Invalid json", ExpectedString: "invalid character '}' looking for beginning of value", Name: "node", File: "package.json", PackageContents: "}"},
		{Case: "Invalid toml", ExpectedString: "toml: line 1: unexpected end of table name (table names cannot be empty)", Name: "cargo", File: "Cargo.toml", PackageContents: "["},
	}

	for _, tc := range cases {
		env, props := getMockedPackageEnv(tc)
		pkg := &Project{}
		pkg.Init(props, env)
		assert.True(t, pkg.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, pkg.Template(), pkg), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
