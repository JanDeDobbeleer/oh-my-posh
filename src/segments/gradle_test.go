package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/alecthomas/assert"
)

const gradleVersionOutput = `------------------------------------------------------------
Gradle 8.12.1
------------------------------------------------------------

Build time:    2025-01-24 12:55:12 UTC
Revision:      0b1ee1ff81d1f4a26574ff4a362ac9180852b140

Kotlin:        2.0.21
Groovy:        3.0.22
Ant:           Apache Ant(TM) version 1.10.15 compiled on August 25 2024
Launcher JVM:  21.0.9 (Eclipse Adoptium 21.0.9+10-LTS)
Daemon JVM:    C:\Users\trajano\scoop\apps\temurin21-jdk\current (no JDK specified, using current Java home)
OS:            Windows 11 10.0 amd64`

func TestGradle(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		GradleOutput   string
		GradlewOutput  string
		ExpectedKotlin string
		ExpectedGroovy string
		ExpectedAnt    string
		ExpectedJVM    string
		HasGradlew     bool
	}{
		{
			Case:           "Global gradle binary",
			ExpectedString: "8.12.1",
			GradleOutput:   gradleVersionOutput,
			HasGradlew:     false,
			ExpectedKotlin: "2.0.21",
			ExpectedGroovy: "3.0.22",
			ExpectedAnt:    "1.10.15",
			ExpectedJVM:    "21.0.9",
		},
		{
			Case:           "Local gradlew wrapper takes priority",
			ExpectedString: "8.12.1",
			GradleOutput:   "",
			GradlewOutput:  gradleVersionOutput,
			HasGradlew:     true,
			ExpectedKotlin: "2.0.21",
			ExpectedGroovy: "3.0.22",
			ExpectedAnt:    "1.10.15",
			ExpectedJVM:    "21.0.9",
		},
		{
			Case:           "Output missing extra version lines",
			ExpectedString: "8.12.1",
			GradleOutput:   "Gradle 8.12.1",
			HasGradlew:     false,
			ExpectedKotlin: "",
			ExpectedGroovy: "",
			ExpectedAnt:    "",
			ExpectedJVM:    "",
		},
	}

	// "No gradle files" — Enabled() returns false without calling gradle at all.
	t.Run("No gradle files in directory", func(t *testing.T) {
		env := new(mock.Environment)
		env.On("HasFiles", "*.gradle").Return(false)
		env.On("HasFiles", "*.gradle.kts").Return(false)
		env.On("HasParentFilePath", "gradlew", false).Return(&runtime.FileInfo{}, errors.New("no match"))
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")

		props := options.Map{options.FetchVersion: true}
		g := &Gradle{}
		g.Init(props, env)
		assert.False(t, g.Enabled())
	})

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.Case, func(t *testing.T) {
			params := &mockedLanguageParams{
				cmd:           "gradle",
				versionParam:  "--version",
				versionOutput: tc.GradleOutput,
				extension:     "*.gradle",
			}
			env, props := getMockedLanguageEnv(params)

			// getMockedLanguageEnv only registers HasFiles for "*.gradle"; also register "*.gradle.kts"
			env.On("HasFiles", "*.gradle.kts").Return(false)
			env.On("Shell").Return("bash")

			fileInfo := &runtime.FileInfo{
				Path:         "../gradlew",
				ParentFolder: "./",
				IsDir:        false,
			}

			var err error
			if !tc.HasGradlew {
				err = errors.New("no match")
			}
			env.On("HasParentFilePath", "gradlew", false).Return(fileInfo, err)

			if tc.HasGradlew {
				env.On("RunCommand", fileInfo.Path, []string{"--version"}).Return(tc.GradlewOutput, nil)
			}

			// Initialize template system so versionURLTemplate rendering doesn't panic.
			if template.Cache == nil {
				template.Cache = &cache.Template{}
			}
			template.Init(env, nil, nil)

			g := &Gradle{}
			g.Init(props, env)
			assert.True(t, g.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, g.Template(), g), fmt.Sprintf("Failed in case: %s", tc.Case))
			assert.Equal(t, tc.ExpectedKotlin, g.KotlinVersion, fmt.Sprintf("Kotlin in case: %s", tc.Case))
			assert.Equal(t, tc.ExpectedGroovy, g.GroovyVersion, fmt.Sprintf("Groovy in case: %s", tc.Case))
			assert.Equal(t, tc.ExpectedAnt, g.AntVersion, fmt.Sprintf("Ant in case: %s", tc.Case))
			assert.Equal(t, tc.ExpectedJVM, g.JVMVersion, fmt.Sprintf("JVM in case: %s", tc.Case))
		})
	}
}
