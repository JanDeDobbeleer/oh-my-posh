package segments

type Deno struct {
	Language
}

func (d *Deno) Template() string {
	return languageTemplate
}

func (d *Deno) Enabled() bool {
	d.extensions = []string{"*.js", "*.ts", "deno.json"}
	d.tooling = map[string]*cmd{
		denoToolName: {
			executable: denoToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegexPrefixed,
		},
	}
	d.defaultTooling = []string{denoToolName}
	d.versionURLTemplate = "https://github.com/denoland/deno/releases/tag/v{{.Full}}"

	return d.Language.Enabled()
}
