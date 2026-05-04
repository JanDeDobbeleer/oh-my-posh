package segments

type Kotlin struct {
	Language
}

func (k *Kotlin) Template() string {
	return languageTemplate
}

const kotlinToolName = "kotlin"

func (k *Kotlin) Enabled() bool {
	k.extensions = []string{"*.kt", "*.kts", "*.ktm"}
	k.tooling = map[string]*cmd{
		kotlinToolName: {
			executable: kotlinToolName,
			args:       []string{versionFlagShortArg},
			regex:      `Kotlin version ` + versionRegex,
		},
	}
	k.defaultTooling = []string{kotlinToolName}
	k.versionURLTemplate = "https://github.com/JetBrains/kotlin/releases/tag/v{{ .Full }}"

	return k.Language.Enabled()
}
