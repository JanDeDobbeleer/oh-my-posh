package segments

type V struct {
	language
}

func (v *V) Template() string {
	return languageTemplate
}

func (v *V) Enabled() bool {
	v.extensions = []string{"*.v"}

	v.commands = []*cmd{
		{
			executable: "v",
			args:       []string{"--version"},
			regex:      `V (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)) [a-f0-9]+`,
		},
	}
	return v.language.Enabled()
}
