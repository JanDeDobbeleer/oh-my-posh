package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// ZigIcon is the icon displayed before the version.
const ZigIcon options.Option = "zigicon"

// Zvm represents the Zig Version Manager (zvm) segment.
type Zvm struct {
	Base

	Version string
	ZigIcon string
}

func (z *Zvm) Template() string {
	return " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} "
}

func (z *Zvm) Enabled() bool {
	if !z.env.HasCommand("zvm") {
		return false
	}

	z.ZigIcon = z.options.String(ZigIcon, "ZVM")

	// Disable colors so the output has no ANSI escape codes to parse.
	output, err := z.env.RunCommand("zvm", "--color=false", "list")
	if err != nil {
		return false
	}

	z.Version = parseActiveZvmVersion(output)

	return z.Version != ""
}

// parseActiveZvmVersion extracts the active version, marked with "[x]", from `zvm list` output.
func parseActiveZvmVersion(output string) string {
	for line := range strings.SplitSeq(output, "\n") {
		if !strings.Contains(line, "[x]") {
			continue
		}

		return strings.TrimSpace(strings.ReplaceAll(line, "[x]", ""))
	}

	return ""
}
