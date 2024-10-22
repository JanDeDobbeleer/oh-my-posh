package segments

type AzFunc struct {
	language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

func (az *AzFunc) Enabled() bool {
	az.extensions = []string{"host.json", "local.settings.json", "function.json"}
	az.commands = []*cmd{
		{

			executable: "func",
			args:       []string{"--version"},
			regex:      `(?P<version>[0-9.]+)`,
		},
	}

	return az.language.Enabled()
}
