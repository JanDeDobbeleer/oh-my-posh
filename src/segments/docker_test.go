package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestDockerSegment(t *testing.T) {
	type envVar struct {
		name  string
		value string
	}
	cases := []struct {
		Case            string
		Expected        string
		ExpectedEnabled bool
		EnvVar          envVar
		HasFiles        bool
		ConfigFile      string
	}{
		{Case: "DOCKER_MACHINE_NAME", Expected: "alpine", ExpectedEnabled: true, EnvVar: envVar{name: "DOCKER_MACHINE_NAME", value: "alpine"}},
		{Case: "DOCKER_HOST", Expected: "alpine 2", ExpectedEnabled: true, EnvVar: envVar{name: "DOCKER_HOST", value: "alpine 2"}},
		{Case: "DOCKER_CONTEXT", Expected: "alpine 3", ExpectedEnabled: true, EnvVar: envVar{name: "DOCKER_HOST", value: "alpine 3"}},
		{Case: "DOCKER_CONTEXT - default", ExpectedEnabled: false, EnvVar: envVar{name: "DOCKER_HOST", value: "default"}},
		{Case: "no docker context active", ExpectedEnabled: false},
		{Case: "config file", Expected: "alpine", ExpectedEnabled: true, HasFiles: true, ConfigFile: `{"currentContext": "alpine"}`},
		{Case: "config file - default", ExpectedEnabled: false, HasFiles: true, ConfigFile: `{"currentContext": "default"}`},
		{Case: "config file - broken", ExpectedEnabled: false, HasFiles: true, ConfigFile: `{`},
	}

	for _, tc := range cases {
		docker := &Docker{}
		env := new(mock.MockedEnvironment)
		docker.Init(properties.Map{}, env)

		for _, v := range docker.envVars() {
			var value string
			if v == tc.EnvVar.name {
				value = tc.EnvVar.value
			}
			env.On("Getenv", v).Return(value)
		}

		env.On("Home").Return("")
		env.On("Getenv", "DOCKER_CONFIG").Return("")
		for _, f := range docker.configFiles() {
			env.On("HasFiles", f).Return(tc.HasFiles)
			env.On("FileContent", f).Return(tc.ConfigFile)
		}

		assert.Equal(t, tc.ExpectedEnabled, docker.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, "{{ .Context }}", docker), tc.Case)
		}
	}
}
