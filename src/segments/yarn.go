package segments

type Yarn struct {
	Language
}

func (n *Yarn) Template() string {
	return " \ue6a7 {{.Full}} "
}

func (n *Yarn) Enabled() bool {
	n.extensions = []string{fileName, "yarn.lock"}
	n.tooling = map[string]*cmd{
		yarnToolName: {
			executable: yarnToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegex,
		},
	}
	n.defaultTooling = []string{yarnToolName}
	n.versionURLTemplate = "https://github.com/yarnpkg/berry/releases/tag/v{{ .Full }}"

	return n.Language.Enabled()
}
