package segments

type XMake struct {
	language
}

func (x *XMake) Template() string {
	return languageTemplate
}

func (x *XMake) Enabled() bool {
	x.extensions = []string{"xmake.lua"}
	x.commands = []*cmd{
		{
			executable: "xmake",
			args:       []string{"--version"},
			regex:      `xmake v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}

	return x.language.Enabled()
}
