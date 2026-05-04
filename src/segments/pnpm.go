package segments

type Pnpm struct {
	Language
}

func (n *Pnpm) Enabled() bool {
	n.extensions = []string{fileName, "pnpm-lock.yaml"}
	n.tooling = map[string]*cmd{
		pnpmToolName: {
			executable: pnpmToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegex,
		},
	}
	n.defaultTooling = []string{pnpmToolName}
	n.versionURLTemplate = "https://github.com/pnpm/pnpm/releases/tag/v{{ .Full }}"

	return n.Language.Enabled()
}

func (n *Pnpm) Template() string {
	return " \ue865 {{.Full}} "
}
