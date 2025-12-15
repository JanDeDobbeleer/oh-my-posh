package segments

type V struct {
	Language
}

func (v *V) Template() string {
	return languageTemplate
}

func (v *V) Enabled() bool {
	v.extensions = []string{"*.v"}

	v.tooling = map[string]*cmd{
		"v": {
			executable: "v",
			args:       []string{"--version"},
			regex:      `V (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)) [a-f0-9]+`,
		},
	}
	v.defaultTooling = []string{"v"}
	return v.Language.Enabled()
}
