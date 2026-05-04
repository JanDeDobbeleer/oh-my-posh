package segments

type R struct {
	Language
}

func (r *R) Template() string {
	return languageTemplate
}

const (
	rscriptToolName = "Rscript"
	rExeToolName    = "R.exe"
)

func (r *R) Enabled() bool {
	rRegex := `version ` + versionRegex
	r.extensions = []string{"*.R", "*.Rmd", "*.Rsx", "*.Rda", "*.Rd", "*.Rproj", ".Rproj.user"}
	r.tooling = map[string]*cmd{
		rscriptToolName: {
			executable: rscriptToolName,
			args:       []string{versionFlagArg},
			regex:      rRegex,
		},
		"R": {
			executable: "R",
			args:       []string{versionFlagArg},
			regex:      rRegex,
		},
		rExeToolName: {
			executable: rExeToolName,
			args:       []string{versionFlagArg},
			regex:      rRegex,
		},
	}
	r.defaultTooling = []string{rscriptToolName, "R", rExeToolName}
	r.versionURLTemplate = "https://www.r-project.org/"

	return r.Language.Enabled()
}
