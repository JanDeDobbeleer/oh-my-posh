package segments

type Swift struct {
	Language
}

func (s *Swift) Template() string {
	return languageTemplate
}

func (s *Swift) Enabled() bool {
	s.extensions = []string{"*.swift", "*.SWIFT", "Podfile"}
	s.tooling = map[string]*cmd{
		"swift": {
			executable: "swift",
			args:       []string{"--version"},
			regex:      `Swift version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)((.|-)(?P<patch>[0-9]+|dev))?))`,
		},
	}
	s.defaultTooling = []string{"swift"}
	s.versionURLTemplate = "https://github.com/apple/swift/releases/tag/swift-{{ .Full }}-RELEASE"

	return s.Language.Enabled()
}
