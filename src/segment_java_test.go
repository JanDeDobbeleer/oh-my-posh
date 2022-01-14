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
			Case:           "Zulu LTS",
			ExpectedString: "11.0.13",
			Version:        "OpenJDK 64-Bit Server VM (11.0.13+8-LTS) for windows-amd64 JRE (Zulu11.52+13-CA) (11.0.13+8-LTS), built on Oct 7 2021 16:00:23 by \"zulu_re\" with MS VC++ 15.9 (VS2017)", // nolint:lll
		},
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
		{
			Case:            "openjdk version \"15.0.2\" 2021-01-19",
			ExpectedString:  "15.0.2",
			JavaHomeEnabled: true,
			JavaHomeVersion: "OpenJDK 64-Bit Server VM (15.0.2+7) for windows-amd64 JRE (15.0.2+7), built on Jan 21 2021 05:54:57 by \"\" with MS VC++ 15.9 (VS2017)",
			Version:         "OpenJDK 64-Bit Server VM (15.0.2+7) for windows-amd64 JRE (15.0.2+7), built on Jan 21 2021 05:54:57 by \"\" with MS VC++ 15.9 (VS2017)",
		},
		{
			Case:            "openjdk version \"16\" 2021-03-16",
			ExpectedString:  "16",
			JavaHomeEnabled: true,
			JavaHomeVersion: "OpenJDK 64-Bit Server VM (16+36) for windows-amd64 JRE (16+36), built on Mar 11 2021 10:56:33 by \"\" with MS VC++ 16.7 (VS2019)",
			Version:         "OpenJDK 64-Bit Server VM (16+36) for windows-amd64 JRE (16+36), built on Mar 11 2021 10:56:33 by \"\" with MS VC++ 16.7 (VS2019)",
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
		env.onTemplate()
		props := properties{
			FetchVersion: true,
		}
		j := &java{}
		j.init(props, env)
		assert.True(t, j.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, j.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
