package segments

import (
	"fmt"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/alecthomas/assert"
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
	env.On("HasFiles", tc.File).Return(true)
	env.On("FileContent", tc.File).Return(tc.PackageContents)
	return env, props
}

func TestPackage(t *testing.T) {
	cases := []*MockData{
		{Case: "1.0.0", ExpectedString: "\uf487 1.0.0", Name: "node", File: "package.json", PackageContents: "{\"version\":\"1.0.0\"}"},
		{Case: "3.2.1", ExpectedString: "\uf487 3.2.1", Name: "node", File: "package.json", PackageContents: "{\"version\":\"3.2.1\"}"},
		{Case: "No version present", ExpectedString: "", Name: "node", File: "package.json", PackageContents: "{}"},
		{Case: "Invalid json", ExpectedString: "invalid character '}' looking for beginning of value", Name: "node", File: "package.json", PackageContents: "}"},
	}

	for _, tc := range cases {
		env, props := getMockedPackageEnv(tc)
		pkg := &Project{}
		pkg.Init(props, env)
		assert.True(t, pkg.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, pkg.Template(), pkg), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
