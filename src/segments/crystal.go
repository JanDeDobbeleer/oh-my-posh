package segments

type Crystal struct {
	Language
}

func (c *Crystal) Template() string {
	return languageTemplate
}

const crystalToolName = "crystal"

func (c *Crystal) Enabled() bool {
	c.extensions = []string{"*.cr", "shard.yml"}
	c.tooling = map[string]*cmd{
		crystalToolName: {
			executable: crystalToolName,
			args:       []string{versionFlagArg},
			regex:      `Crystal ` + versionRegex,
		},
	}
	c.defaultTooling = []string{crystalToolName}
	c.versionURLTemplate = "https://github.com/crystal-lang/crystal/releases/tag/{{ .Full }}"

	return c.Language.Enabled()
}
