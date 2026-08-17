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
		// Not marked versionCacheable: on macOS /usr/bin/java is Apple's
		// stub, resolving the actual JDK via java_home at run time - the
		// stub's identity stays fixed while the reported version follows
		// whichever JDK is installed/selected. The java_home cmd below stays
		// cacheable because there the resolved path itself keys correctly.
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
			// JAVA_HOME is baked directly into executable above, so a
			// different JAVA_HOME resolves to a different path and therefore
			// a different cache key automatically - the env-var dependency
			// is fully captured by the resolved path, unlike a tool that
			// reads JAVA_HOME again at run time.
			versionCacheable: true,
		}
		j.defaultTooling = []string{"java_home", javaToolName}
	}
}
