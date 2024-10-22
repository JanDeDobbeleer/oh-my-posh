package segments

import (
	"fmt"
)

type Java struct {
	language
}

func (j *Java) Template() string {
	return languageTemplate
}

func (j *Java) Enabled() bool {
	j.init()

	return j.language.Enabled()
}

func (j *Java) init() {
	javaRegex := `(?: JRE)(?: \(.*\))? \((?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))?(?:\.(?P<patch>[0-9]+))?).*\),`
	javaCmd := &cmd{
		executable: "java",
		args:       []string{"-Xinternalversion"},
		regex:      javaRegex,
	}

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

	javaHome := j.env.Getenv("JAVA_HOME")
	if len(javaHome) > 0 {
		java := fmt.Sprintf("%s/bin/java", javaHome)
		j.commands = []*cmd{
			{
				executable: java,
				args:       []string{"-Xinternalversion"},
				regex:      javaRegex,
			},
			javaCmd,
		}
		return
	}

	j.commands = []*cmd{javaCmd}
}
