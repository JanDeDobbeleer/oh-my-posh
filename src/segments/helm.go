package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Helm struct {
	props properties.Properties
	env   platform.Environment

	Version string
}

func (h *Helm) Enabled() bool {
	displayMode := h.props.GetString(DisplayMode, DisplayModeAlways)
	if displayMode != DisplayModeFiles {
		return h.getVersion()
	}

	inChart := false
	files := []string{"Chart.yml", "Chart.yaml"}
	for _, file := range files {
		if _, err := h.env.HasParentFilePath(file); err == nil {
			inChart = true
			break
		}
	}

	return inChart && h.getVersion()
}

func (h *Helm) Template() string {
	return " Helm {{.Version}}"
}

func (h *Helm) Init(props properties.Properties, env platform.Environment) {
	h.props = props
	h.env = env
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
