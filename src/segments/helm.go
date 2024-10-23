package segments

type Helm struct {
	base

	Version string
}

func (h *Helm) Enabled() bool {
	displayMode := h.props.GetString(DisplayMode, DisplayModeAlways)
	if displayMode != DisplayModeFiles {
		return h.getVersion()
	}

	inChart := false
	files := []string{"Chart.yml", "Chart.yaml", "helmfile.yaml", "helmfile.yml"}
	for _, file := range files {
		if _, err := h.env.HasParentFilePath(file, false); err == nil {
			inChart = true
			break
		}
	}

	return inChart && h.getVersion()
}

func (h *Helm) Template() string {
	return " Helm {{.Version}}"
}

func (h *Helm) getVersion() bool {
	cmd := "helm"
	if !h.env.HasCommand(cmd) {
		return false
	}

	result, err := h.env.RunCommand(cmd, "version", "--short", "--template={{.Version}}")
	if err != nil {
		return false
	}

	h.Version = result[1:]
	return true
}
