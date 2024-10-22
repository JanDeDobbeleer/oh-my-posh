package segments

type R struct {
	language
}

func (r *R) Template() string {
	return languageTemplate
}

func (r *R) Enabled() bool {
	rRegex := `version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`
	r.extensions = []string{"*.R", "*.Rmd", "*.Rsx", "*.Rda", "*.Rd", "*.Rproj", ".Rproj.user"}
	r.commands = []*cmd{
		{
			executable: "Rscript",
			args:       []string{"--version"},
			regex:      rRegex,
		},
		{
			executable: "R",
			args:       []string{"--version"},
			regex:      rRegex,
		},
		{
			executable: "R.exe",
			args:       []string{"--version"},
			regex:      rRegex,
		},
	}
	r.versionURLTemplate = "https://www.r-project.org/"

	return r.language.Enabled()
}
