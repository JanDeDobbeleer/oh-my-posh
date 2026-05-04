package segments

type AzFunc struct {
	Language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

const azFuncToolName = "func"

func (az *AzFunc) Enabled() bool {
	az.extensions = []string{"host.json", "local.settings.json", "function.json"}
	az.tooling = map[string]*cmd{
		azFuncToolName: {
			executable: azFuncToolName,
			args:       []string{versionFlagArg},
			regex:      `(?P<version>[0-9.]+)`,
		},
	}
	az.defaultTooling = []string{azFuncToolName}

	return az.Language.Enabled()
}
