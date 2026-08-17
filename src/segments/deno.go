package segments

type Deno struct {
	Language
}

func (d *Deno) Template() string {
	return languageTemplate
}

func (d *Deno) Enabled() bool {
	d.loadSpec()

	return d.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (d *Deno) Activation() Activation {
	d.loadSpec()

	return d.activation()
}

func (d *Deno) loadSpec() {
	d.extensions = []string{"*.js", "*.ts", "deno.json"}
	d.tooling = map[string]*cmd{
		denoToolName: {
			executable:       denoToolName,
			args:             []string{versionFlagArg},
			regex:            versionRegexPrefixed,
			versionCacheable: true,
		},
	}
	d.defaultTooling = []string{denoToolName}
	d.versionURLTemplate = "https://github.com/denoland/deno/releases/tag/v{{.Full}}"
}
