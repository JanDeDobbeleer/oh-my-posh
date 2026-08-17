package segments

type Pnpm struct {
	Language
}

func (n *Pnpm) Enabled() bool {
	n.loadSpec()

	return n.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (n *Pnpm) Activation() Activation {
	n.loadSpec()

	return n.activation()
}

func (n *Pnpm) loadSpec() {
	n.extensions = []string{fileName, "pnpm-lock.yaml"}
	n.tooling = map[string]*cmd{
		// Not marked versionCacheable: same Corepack shim risk as yarn (see
		// yarn.go) - pnpm is equally commonly managed through Corepack's
		// package.json-pinned dispatch.
		pnpmToolName: {
			executable: pnpmToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegex,
		},
	}
	n.defaultTooling = []string{pnpmToolName}
	n.versionURLTemplate = "https://github.com/pnpm/pnpm/releases/tag/v{{ .Full }}"
}

func (n *Pnpm) Template() string {
	return " \ue865 {{.Full}} "
}
