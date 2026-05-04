package segments

type Clojure struct {
	Language
}

func (c *Clojure) Template() string {
	return languageTemplate
}

func (c *Clojure) Enabled() bool {
	c.init()
	return c.Language.Enabled()
}

const clojureToolName = "clojure"

const leinToolName = "lein"

func (c *Clojure) init() {
	options := c.options.StringArray(Tooling, []string{})
	if len(options) == 0 {
		c.defaultTooling = []string{clojureToolName, leinToolName}
	}

	c.extensions = []string{
		"project.clj",
		"deps.edn",
		"build.boot",
		"bb.edn",
		"*.clj",
		"*.cljc",
		"*.cljs",
	}

	c.tooling = map[string]*cmd{
		clojureToolName: {
			executable: clojureToolName,
			args:       []string{versionFlagArg},
			regex:      `Clojure CLI version (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+)(?:\.(?P<build>[0-9]+))?)`,
		},
		leinToolName: {
			executable: leinToolName,
			args:       []string{versionFlagArg},
			regex:      `Leiningen (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+))`,
		},
	}
}
