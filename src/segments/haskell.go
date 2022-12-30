package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
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

func (h *Haskell) Init(props properties.Properties, env platform.Environment) {
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

	h.language = language{
		env:                env,
		props:              props,
		extensions:         []string{"*.hs", "*.lhs", "stack.yaml", "package.yaml", "*.cabal", "cabal.project"},
		commands:           []*cmd{ghcCmd},
		versionURLTemplate: "https://www.haskell.org/ghc/download_ghc_{{ .Major }}_{{ .Minor }}_{{ .Patch }}.html",
	}

	switch h.props.GetString(StackGhcMode, "never") {
	case "always":
		h.language.commands = []*cmd{stackGhcCmd}
		h.StackGhc = true
	case "package":
		_, err := h.language.env.HasParentFilePath("stack.yaml")
		if err == nil {
			h.language.commands = []*cmd{stackGhcCmd}
			h.StackGhc = true
		}
	}
}

func (h *Haskell) Enabled() bool {
	return h.language.Enabled()
}
