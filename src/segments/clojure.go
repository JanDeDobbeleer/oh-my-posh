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

func (c *Clojure) init() {
	options := c.options.StringArray(Tooling, []string{})
	if len(options) == 0 {
		c.defaultTooling = []string{"clojure", "lein"}
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
		"clojure": {
			executable: "clojure",
			args:       []string{"--version"},
			regex:      `Clojure CLI version (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+)(?:\.(?P<build>[0-9]+))?)`,
		},
		"lein": {
			executable: "lein",
			args:       []string{"--version"},
			regex:      `Leiningen (?P<version>(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+))`,
		},
	}
}
