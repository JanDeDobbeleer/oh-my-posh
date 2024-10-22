package segments

type Deno struct {
	language
}

func (d *Deno) Template() string {
	return languageTemplate
}

func (d *Deno) Enabled() bool {
	d.extensions = []string{"*.js", "*.ts", "deno.json"}
	d.commands = []*cmd{
		{
			executable: "deno",
			args:       []string{"--version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	d.versionURLTemplate = "https://github.com/denoland/deno/releases/tag/v{{.Full}}"

	return d.language.Enabled()
}
