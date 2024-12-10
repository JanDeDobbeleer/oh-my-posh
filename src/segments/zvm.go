package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Zvm struct {
	props properties.Properties
	env   runtime.Environment

	Version string
	ZigIcon string
}

func (z *Zvm) SetText(text string) {
	z.Version = text
}

func (z *Zvm) Text() string {
	return z.Version
}

const (
	ZvmIcon properties.Property = "zigicon"
)

func (z *Zvm) Enabled() bool {
	if !z.env.HasCommand("zvm") {
		return false
	}
	version := z.getZvmVersion()
	return version != ""
}

func (z *Zvm) Template() string {
	return " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} "
}

func (z *Zvm) Init(props properties.Properties, env runtime.Environment) {
	z.props = props
	z.env = env

	// Initialize the icon from properties
	z.ZigIcon = z.props.GetString(ZvmIcon, "ZVM")
	z.Version = z.getZvmVersion()
}

func (z *Zvm) getZvmVersion() string {
	output, err := z.env.RunCommand("zvm", "list")
	if err != nil {
		return ""
	}

	// Split output into lines
	lines := strings.Split(output, "\n")

	// Look for line containing green color code which indicates active version
	// ANSI color code for green is typically \033[32m or \x1b[32m
	for _, line := range lines {
		if strings.Contains(line, "\x1b[32m") || strings.Contains(line, "\033[32m") {
			// Clean ANSI color codes and whitespace
			cleaned := strings.ReplaceAll(line, "\x1b[32m", "")
			cleaned = strings.ReplaceAll(cleaned, "\033[32m", "")
			cleaned = strings.ReplaceAll(cleaned, "\x1b[0m", "")
			cleaned = strings.TrimSpace(cleaned)

			return cleaned
		}
	}
	return ""
}
