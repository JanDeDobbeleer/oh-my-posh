package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Gradle struct {
	Language
	KotlinVersion string
	GroovyVersion string
	AntVersion    string
	JVMVersion    string
}

func (g *Gradle) Template() string {
	return languageTemplate
}

func (g *Gradle) Enabled() bool {
	g.extensions = []string{"*.gradle", "*.gradle.kts"}

	executable := "gradle"
	gradlew, err := g.env.HasParentFilePath("gradlew", false)
	if err == nil {
		executable = gradlew.Path
	}

	g.tooling = map[string]*cmd{
		"gradle": {
			executable: executable,
			args:       []string{"--version"},
			regex:      `Gradle (?P<version>(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)`,
			getVersion: g.buildGetVersion(executable),
		},
	}
	g.defaultTooling = []string{"gradle"}
	g.versionURLTemplate = "https://github.com/gradle/gradle/releases/tag/v{{ .Full }}"

	return g.Language.Enabled()
}

func (g *Gradle) buildGetVersion(executable string) getVersion {
	return func() (string, error) {
		output, err := g.env.RunCommand(executable, "--version")
		if err != nil {
			return "", err
		}
		g.parseExtraVersions(output)
		return output, nil
	}
}

func (g *Gradle) parseExtraVersions(output string) {
	if v := regex.FindNamedRegexMatch(`Kotlin:\s+(?P<version>\S+)`, output); len(v) > 0 {
		g.KotlinVersion = v["version"]
	}
	if v := regex.FindNamedRegexMatch(`Groovy:\s+(?P<version>\S+)`, output); len(v) > 0 {
		g.GroovyVersion = v["version"]
	}
	if v := regex.FindNamedRegexMatch(`Apache Ant.*version (?P<version>[\d.]+)`, output); len(v) > 0 {
		g.AntVersion = v["version"]
	}
	if v := regex.FindNamedRegexMatch(`Launcher JVM:\s+(?P<version>[\d._]+)`, output); len(v) > 0 {
		g.JVMVersion = v["version"]
	}
}
