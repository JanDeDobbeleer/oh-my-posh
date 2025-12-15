package segments

type XMake struct {
	Language
}

func (x *XMake) Template() string {
	return languageTemplate
}

func (x *XMake) Enabled() bool {
	x.extensions = []string{"xmake.lua"}
	x.tooling = map[string]*cmd{
		"xmake": {
			executable: "xmake",
			args:       []string{"--version"},
			regex:      `xmake v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	x.defaultTooling = []string{"xmake"}

	return x.Language.Enabled()
}
