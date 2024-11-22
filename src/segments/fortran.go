package segments

type Fortran struct {
	language
}

func (f *Fortran) Template() string {
	return languageTemplate
}

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
	f.commands = []*cmd{
		{
			executable: "gfortran",
			args:       []string{"--version"},
			regex:      `GNU Fortran \(.*\) (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}

	return f.language.Enabled()
}
