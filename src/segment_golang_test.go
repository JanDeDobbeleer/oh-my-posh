package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGolang(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Go 1.15", ExpectedString: "1.15.8", Version: "go version go1.15.8 darwin/amd64"},
		{Case: "Go 1.16", ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "go").Return(true)
		env.On("runCommand", "go", []string{"version"}).Return(tc.Version, nil)
		env.On("hasFiles", "*.go").Return(true)
		env.On("getcwd", nil).Return("/usr/home/project")
		env.On("homeDir", nil).Return("/usr/home")
		props := &properties{
			values: map[Property]interface{}{
				DisplayVersion: true,
			},
		}
		g := &golang{}
		g.init(props, env)
		assert.True(t, g.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, g.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
