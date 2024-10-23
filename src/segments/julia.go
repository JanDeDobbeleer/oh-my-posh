package segments

type Julia struct {
	language
}

func (j *Julia) Template() string {
	return languageTemplate
}

func (j *Julia) Enabled() bool {
	j.extensions = []string{"*.jl"}
	j.commands = []*cmd{
		{
			executable: "julia",
			args:       []string{"--version"},
			regex:      `julia version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	j.versionURLTemplate = "https://github.com/JuliaLang/julia/releases/tag/v{{ .Full }}"

	return j.language.Enabled()
}
