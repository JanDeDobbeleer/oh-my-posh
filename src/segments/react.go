package segments

type React struct {
	language
}

func (r *React) Template() string {
	return languageTemplate
}

func (r *React) Enabled() bool {
	r.extensions = []string{"package.json"}
	r.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: r.getVersion,
		},
	}
	r.versionURLTemplate = "https://github.com/facebook/react/releases/tag/v{{.Full}}"

	return r.language.Enabled()
}

func (r *React) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(r.language.env, "react")
}
