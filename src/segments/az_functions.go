package segments

type AzFunc struct {
	Language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

func (az *AzFunc) Enabled() bool {
	az.extensions = []string{"host.json", "local.settings.json", "function.json"}
	az.tooling = map[string]*cmd{
		"func": {
			executable: "func",
			args:       []string{"--version"},
			regex:      `(?P<version>[0-9.]+)`,
		},
	}
	az.defaultTooling = []string{"func"}

	return az.Language.Enabled()
}
