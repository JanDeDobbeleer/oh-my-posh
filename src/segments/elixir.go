package segments

type Elixir struct {
	Language
}

func (e *Elixir) Template() string {
	return languageTemplate
}

func (e *Elixir) Enabled() bool {
	e.extensions = []string{"*.ex", "*.exs"}
	e.tooling = map[string]*cmd{
		"asdf": {
			executable: "asdf",
			args:       []string{"current", "elixir"},
			regex:      `elixir\s+(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))[^\s]*\s+`,
		},
		"elixir": {
			executable: "elixir",
			args:       []string{"--version"},
			regex:      `Elixir (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	e.defaultTooling = []string{"asdf", "elixir"}
	e.versionURLTemplate = "https://github.com/elixir-lang/elixir/releases/tag/v{{ .Full }}"

	return e.Language.Enabled()
}
