package segments

import (
	"fmt"
)

type Java struct {
	Language
}

func (j *Java) Template() string {
	return languageTemplate
}

func (j *Java) Enabled() bool {
	j.loadSpec()

	return j.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (j *Java) Activation() Activation {
	j.loadSpec()

	return j.activation()
}

const javaToolName = "java"

func (j *Java) loadSpec() {
	javaRegex := `(?: JRE)(?: \(.*\))? \((?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))?(?:\.(?P<patch>[0-9]+))?).*\),`

	j.extensions = []string{
		"pom.xml",
		"build.gradle.kts",
		"build.sbt",
		".java-version",
		".deps.edn",
		"project.clj",
		"build.boot",
		"*.java",
		"*.class",
		"*.gradle",
		"*.jar",
		"*.clj",
		"*.cljc",
	}

	j.tooling = map[string]*cmd{
		javaToolName: {
			executable: javaToolName,
			args:       []string{"-Xinternalversion"},
			regex:      javaRegex,
		},
	}
	j.defaultTooling = []string{javaToolName}

	javaHome := j.env.Getenv("JAVA_HOME")
	if len(javaHome) > 0 {
		java := fmt.Sprintf("%s/bin/java", javaHome)
		j.tooling["java_home"] = &cmd{
			executable: java,
			args:       []string{"-Xinternalversion"},
			regex:      javaRegex,
		}
		j.defaultTooling = []string{"java_home", javaToolName}
	}
}
