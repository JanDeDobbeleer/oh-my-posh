package segments

type Kotlin struct {
	language
}

func (k *Kotlin) Template() string {
	return languageTemplate
}

func (k *Kotlin) Enabled() bool {
	k.extensions = []string{"*.kt", "*.kts", "*.ktm"}
	k.commands = []*cmd{
		{
			executable: "kotlin",
			args:       []string{"-version"},
			regex:      `Kotlin version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	k.versionURLTemplate = "https://github.com/JetBrains/kotlin/releases/tag/v{{ .Full }}"

	return k.language.Enabled()
}
