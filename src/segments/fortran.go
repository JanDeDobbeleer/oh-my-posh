package segments

type Fortran struct {
	Language
}

func (f *Fortran) Template() string {
	return languageTemplate
}

const gfortranToolName = "gfortran"

func (f *Fortran) Enabled() bool {
	f.extensions = []string{
		"*.f", "*.for", "*.fpp",
		"*.f77", "*.f90", "*.f95",
		"*.f03", "*.f08",
		"*.F", "*.FOR", "*.FPP",
		"*.F77", "*.F90", "*.F95",
		"*.F03", "*.F08",
		"fpm.toml",
	}
	f.tooling = map[string]*cmd{
		gfortranToolName: {
			executable: gfortranToolName,
			args:       []string{versionFlagArg},
			regex:      `GNU Fortran \(.*\) ` + versionRegex,
		},
	}
	f.defaultTooling = []string{gfortranToolName}

	return f.Language.Enabled()
}
