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
		// Not marked versionCacheable: when Corepack manages yarn (the
		// officially recommended setup for yarn >=2), the "yarn" resolved
		// from PATH is a Corepack shim whose own file never changes but which
		// dispatches to whatever version the current project's package.json
		// "packageManager" field pins - so the same resolved path/mtime/size
		// can legitimately report different versions in different projects.
		yarnToolName: {
			executable: yarnToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegex,
		},
	}
	n.defaultTooling = []string{yarnToolName}
	n.versionURLTemplate = "https://github.com/yarnpkg/berry/releases/tag/v{{ .Full }}"
}
