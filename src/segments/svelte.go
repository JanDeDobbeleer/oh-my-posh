package segments

type Svelte struct {
	Language
}

func (s *Svelte) Template() string {
	return languageTemplate
}

func (s *Svelte) Enabled() bool {
	s.loadSpec()

	return s.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (s *Svelte) Activation() Activation {
	s.loadSpec()

	return s.activation()
}

func (s *Svelte) loadSpec() {
	s.extensions = []string{"svelte.config.js"}
	s.tooling = map[string]*cmd{
		"svelte": {
			regex:      versionRegexPrefixed,
			getVersion: s.getVersion,
		},
	}
	s.defaultTooling = []string{"svelte"}
	s.versionURLTemplate = "https://github.com/sveltejs/svelte/releases/tag/svelte%40{{.Full}}"
}

func (s *Svelte) getVersion() (string, error) {
	return s.nodePackageVersion("svelte")
}
