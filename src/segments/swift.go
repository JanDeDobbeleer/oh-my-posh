package segments

type Swift struct {
	language
}

func (s *Swift) Template() string {
	return languageTemplate
}

func (s *Swift) Enabled() bool {
	s.extensions = []string{"*.swift", "*.SWIFT", "Podfile"}
	s.commands = []*cmd{
		{
			executable: "swift",
			args:       []string{"--version"},
			regex:      `Swift version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)((.|-)(?P<patch>[0-9]+|dev))?))`,
		},
	}
	s.versionURLTemplate = "https://github.com/apple/swift/releases/tag/swift-{{ .Full }}-RELEASE"

	return s.language.Enabled()
}
