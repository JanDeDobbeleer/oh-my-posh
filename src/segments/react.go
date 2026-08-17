//nolint:dupl // react and aurelia are deliberately parallel: identical node-package detection, differing only in package name and version metadata
package segments

type React struct {
	Language
}

func (r *React) Template() string {
	return languageTemplate
}

func (r *React) Enabled() bool {
	r.loadSpec()

	if !r.hasNodePackage("react") {
		return false
	}

	return r.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (r *React) Activation() Activation {
	r.loadSpec()

	return r.activation()
}

func (r *React) loadSpec() {
	r.extensions = []string{fileName}
	r.tooling = map[string]*cmd{
		"react": {
			regex:      versionRegexPrefixed,
			getVersion: r.getVersion,
		},
	}
	r.defaultTooling = []string{"react"}
	r.versionURLTemplate = "https://github.com/facebook/react/releases/tag/v{{.Full}}"
}

func (r *React) getVersion() (string, error) {
	return r.nodePackageVersion("react")
}
