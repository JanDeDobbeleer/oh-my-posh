package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJava(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		Version         string
		JavaHomeVersion string
		JavaHomeEnabled bool
	}{
		{
			Case:           "OpenJDK macOS",
			ExpectedString: "1.8.0",
			Version:        "OpenJDK 64-Bit Server VM (25.275-b01) for bsd-amd64 JRE (1.8.0_275-b01), built on Nov  9 2020 12:07:35 by \"jenkins\" with gcc 4.2.1",
		},
		{
			Case:           "OpenJDK macOS with JAVA_HOME, no executable",
			ExpectedString: "1.8.0",
			Version:        "OpenJDK 64-Bit Server VM (25.275-b01) for bsd-amd64 JRE (1.8.0_275-b01), built on Nov  9 2020 12:07:35 by \"jenkins\" with gcc 4.2.1",
		},
		{
			Case:            "OpenJDK macOS with JAVA_HOME and executable",
			ExpectedString:  "1.7.0",
			JavaHomeEnabled: true,
			JavaHomeVersion: "OpenJDK 64-Bit Server VM (25.275-b01) for bsd-amd64 JRE (1.7.0_275-b01), built on Nov  9 2020 12:07:35 by \"jenkins\" with gcc 4.2.1",
			Version:         "OpenJDK 64-Bit Server VM (25.275-b01) for bsd-amd64 JRE (1.8.0_275-b01), built on Nov  9 2020 12:07:35 by \"jenkins\" with gcc 4.2.1",
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "java").Return(true)
		env.On("runCommand", "java", []string{"-Xinternalversion"}).Return(tc.Version, nil)
		env.On("hasFiles", "pom.xml").Return(true)
		env.On("getcwd", nil).Return("/usr/home/project")
		env.On("homeDir", nil).Return("/usr/home")
		if tc.JavaHomeEnabled {
			env.On("getenv", "JAVA_HOME").Return("/usr/java")
			env.On("hasCommand", "/usr/java/bin/java").Return(true)
			env.On("runCommand", "/usr/java/bin/java", []string{"-Xinternalversion"}).Return(tc.JavaHomeVersion, nil)
		} else {
			env.On("getenv", "JAVA_HOME").Return("")
		}
		props := &properties{
			values: map[Property]interface{}{
				DisplayVersion: true,
			},
		}
		j := &java{}
		j.init(props, env)
		assert.True(t, j.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, j.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
