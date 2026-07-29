package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const ZigIcon options.Option = "zigicon"

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

// The active version is marked with "[x]" in `zvm list` output.
func parseActiveZvmVersion(output string) string {
	for line := range strings.SplitSeq(output, "\n") {
		if !strings.Contains(line, "[x]") {
			continue
		}

		return strings.TrimSpace(strings.ReplaceAll(line, "[x]", ""))
	}

	return ""
}
