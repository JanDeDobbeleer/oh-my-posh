package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

const (
	gradle = "gradle"
)

type Gradle struct {
	KotlinVersion string
	GroovyVersion string
	AntVersion    string
	JVMVersion    string
	Language
}

func (g *Gradle) Template() string {
	return languageTemplate
}

func (g *Gradle) Enabled() bool {
	g.loadSpec()

	executable := gradle
	gradlew, err := g.env.HasParentFilePath("gradlew", false)
	if err == nil {
		executable = gradlew.Path
	}

	// Not marked versionCacheable: getVersion is always set below, so this
	// command never takes runCommand's executable-invocation branch (the
	// only one the flag affects) regardless of whether one is set. Left
	// unset anyway to document why - the wrapper (gradlew) case this
	// executable can resolve to reports a version pinned by the project's
	// gradle-wrapper.properties, not by the script's own identity.
	g.tooling = map[string]*cmd{
		gradle: {
			executable: executable,
			args:       []string{"--version"},
			regex:      `Gradle (?P<version>(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)`,
			getVersion: g.buildGetVersion(executable),
		},
	}
	g.defaultTooling = []string{gradle}
	g.versionURLTemplate = "https://github.com/gradle/gradle/releases/tag/v{{ .Full }}"

	return g.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (g *Gradle) Activation() Activation {
	g.loadSpec()

	return g.activation()
}

// loadSpec holds the file spec only; the tooling depends on a gradlew lookup
// that stays in Enabled().
func (g *Gradle) loadSpec() {
	g.extensions = []string{"*.gradle", "*.gradle.kts"}
}

func (g *Gradle) buildGetVersion(executable string) getVersion {
	return func() (string, error) {
		output, err := g.env.RunCommandWithEnv(executable, nil, "--version")
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
