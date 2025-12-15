package segments

type R struct {
	Language
}

func (r *R) Template() string {
	return languageTemplate
}

func (r *R) Enabled() bool {
	rRegex := `version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`
	r.extensions = []string{"*.R", "*.Rmd", "*.Rsx", "*.Rda", "*.Rd", "*.Rproj", ".Rproj.user"}
	r.tooling = map[string]*cmd{
		"Rscript": {
			executable: "Rscript",
			args:       []string{"--version"},
			regex:      rRegex,
		},
		"R": {
			executable: "R",
			args:       []string{"--version"},
			regex:      rRegex,
		},
		"R.exe": {
			executable: "R.exe",
			args:       []string{"--version"},
			regex:      rRegex,
		},
	}
	r.defaultTooling = []string{"Rscript", "R", "R.exe"}
	r.versionURLTemplate = "https://www.r-project.org/"

	return r.Language.Enabled()
}
