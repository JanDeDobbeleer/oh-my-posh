package segments

type Elixir struct {
	Language
}

func (e *Elixir) Template() string {
	return languageTemplate
}

const elixirToolName = "elixir"

func (e *Elixir) Enabled() bool {
	e.extensions = []string{"*.ex", "*.exs"}
	e.tooling = map[string]*cmd{
		asdfToolName: {
			executable: asdfToolName,
			args:       []string{"current", elixirToolName},
			regex:      `elixir\s+` + versionRegex + `[^\s]*\s+`,
		},
		elixirToolName: {
			executable: elixirToolName,
			args:       []string{versionFlagArg},
			regex:      `Elixir ` + versionRegex,
		},
	}
	e.defaultTooling = []string{asdfToolName, elixirToolName}
	e.versionURLTemplate = "https://github.com/elixir-lang/elixir/releases/tag/v{{ .Full }}"

	return e.Language.Enabled()
}
