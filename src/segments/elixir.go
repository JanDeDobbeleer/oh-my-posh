package segments

type Elixir struct {
	language
}

func (e *Elixir) Template() string {
	return languageTemplate
}

func (e *Elixir) Enabled() bool {
	e.extensions = []string{"*.ex", "*.exs"}
	e.commands = []*cmd{
		{
			executable: "asdf",
			args:       []string{"current", "elixir"},
			regex:      `elixir\s+(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))[^\s]*\s+`,
		},
		{
			executable: "elixir",
			args:       []string{"--version"},
			regex:      `Elixir (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	e.versionURLTemplate = "https://github.com/elixir-lang/elixir/releases/tag/v{{ .Full }}"

	return e.language.Enabled()
}
