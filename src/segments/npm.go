package segments

type Npm struct {
	Language
}

func (n *Npm) Enabled() bool {
	n.loadSpec()

	return n.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (n *Npm) Activation() Activation {
	n.loadSpec()

	return n.activation()
}

func (n *Npm) loadSpec() {
	n.extensions = []string{fileName, "package-lock.json"}
	n.tooling = map[string]*cmd{
		// Unlike yarn/pnpm, npm ships bundled with node itself and Corepack
		// does not shim it by default, so npm --version is not routed through
		// a project-pinned dispatcher: it reports the resolved binary's own
		// version.
		npmToolName: {
			executable:       npmToolName,
			args:             []string{versionFlagArg},
			regex:            versionRegex,
			versionCacheable: true,
		},
	}
	n.defaultTooling = []string{npmToolName}
	n.versionURLTemplate = "https://github.com/npm/cli/releases/tag/v{{ .Full }}"
}

func (n *Npm) Template() string {
	return " \ue71e {{.Full}} "
}
