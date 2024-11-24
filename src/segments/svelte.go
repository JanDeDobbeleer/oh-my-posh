package segments

type Svelte struct {
	language
}

func (s *Svelte) Template() string {
	return languageTemplate
}

func (s *Svelte) Enabled() bool {
	s.extensions = []string{"svelte.config.js"}
	s.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			getVersion: s.getVersion,
		},
	}
	s.versionURLTemplate = "https://github.com/sveltejs/svelte/releases/tag/svelte%40{{.Full}}"

	return s.language.Enabled()
}

func (s *Svelte) getVersion() (string, error) {
	return s.nodePackageVersion("svelte")
}
