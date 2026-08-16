package segments

type Yarn struct {
	Language
}

func (n *Yarn) Template() string {
	return " \ue6a7 {{.Full}} "
}

func (n *Yarn) Enabled() bool {
	n.loadSpec()

	return n.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (n *Yarn) Activation() Activation {
	n.loadSpec()

	return n.activation()
}

func (n *Yarn) loadSpec() {
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
}
