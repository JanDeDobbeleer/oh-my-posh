package segments

type Npm struct {
	Language
}

func (n *Npm) Enabled() bool {
	n.extensions = []string{fileName, "package-lock.json"}
	n.tooling = map[string]*cmd{
		npmToolName: {
			executable: npmToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegex,
		},
	}
	n.defaultTooling = []string{npmToolName}
	n.versionURLTemplate = "https://github.com/npm/cli/releases/tag/v{{ .Full }}"

	return n.Language.Enabled()
}

func (n *Npm) Template() string {
	return " \ue71e {{.Full}} "
}
