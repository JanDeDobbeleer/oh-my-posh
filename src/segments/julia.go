package segments

type Julia struct {
	Language
}

func (j *Julia) Template() string {
	return languageTemplate
}

func (j *Julia) Enabled() bool {
	j.extensions = []string{"*.jl"}
	j.tooling = map[string]*cmd{
		juliaToolName: {
			executable: juliaToolName,
			args:       []string{versionFlagArg},
			regex:      `julia version ` + versionRegex,
		},
	}
	j.defaultTooling = []string{juliaToolName}
	j.versionURLTemplate = "https://github.com/JuliaLang/julia/releases/tag/v{{ .Full }}"

	return j.Language.Enabled()
}
