package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Haskell struct {
	language

	StackGhc bool
}

const (
	StackGhcMode properties.Property = "stack_ghc_mode"
)

func (h *Haskell) Template() string {
	return languageTemplate
}

func (h *Haskell) Enabled() bool {
	ghcRegex := `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`
	ghcCmd := &cmd{
		executable: "ghc",
		args:       []string{"--numeric-version"},
		regex:      ghcRegex,
	}

	stackGhcCmd := &cmd{
		executable: "stack",
		args:       []string{"ghc", "--", "--numeric-version"},
		regex:      ghcRegex,
	}

	h.extensions = []string{"*.hs", "*.lhs", "stack.yaml", "package.yaml", "*.cabal", "cabal.project"}
	h.commands = []*cmd{ghcCmd}
	h.versionURLTemplate = "https://www.haskell.org/ghc/download_ghc_{{ .Major }}_{{ .Minor }}_{{ .Patch }}.html"

	switch h.props.GetString(StackGhcMode, "never") {
	case "always":
		h.commands = []*cmd{stackGhcCmd}
		h.StackGhc = true
	case "package":
		_, err := h.env.HasParentFilePath("stack.yaml", false)
		if err == nil {
			h.commands = []*cmd{stackGhcCmd}
			h.StackGhc = true
		}
	}

	return h.language.Enabled()
}
