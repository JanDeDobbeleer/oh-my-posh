package main

import "fmt"

type java struct {
	language
}

func (j *java) template() string {
	return languageTemplate
}

func (j *java) init(props Properties, env Environment) {
	javaRegex := `(?: JRE)(?: \(.*\))? \((?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))?(?:\.(?P<patch>[0-9]+))?).*\),`
	javaCmd := &cmd{
		executable: "java",
		args:       []string{"-Xinternalversion"},
		regex:      javaRegex,
	}
	j.language = language{
		env:   env,
		props: props,
		extensions: []string{
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
		},
	}
	javaHome := j.language.env.Getenv("JAVA_HOME")
	if len(javaHome) > 0 {
		java := fmt.Sprintf("%s/bin/java", javaHome)
		j.language.commands = []*cmd{
			{
				executable: java,
				args:       []string{"-Xinternalversion"},
				regex:      javaRegex,
			},
			javaCmd,
		}
		return
	}
	j.language.commands = []*cmd{javaCmd}
}

func (j *java) enabled() bool {
	return j.language.enabled()
}
