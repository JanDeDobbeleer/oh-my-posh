package segments

type AzFunc struct {
	Language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

const azFuncToolName = "func"

func (az *AzFunc) Enabled() bool {
	az.loadSpec()

	return az.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (az *AzFunc) Activation() Activation {
	az.loadSpec()

	return az.activation()
}

func (az *AzFunc) loadSpec() {
	az.extensions = []string{"host.json", "local.settings.json", "function.json"}
	az.tooling = map[string]*cmd{
		azFuncToolName: {
			executable: azFuncToolName,
			args:       []string{versionFlagArg},
			regex:      `(?P<version>[0-9.]+)`,
		},
	}
	az.defaultTooling = []string{azFuncToolName}
}
