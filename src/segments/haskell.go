package segments

import "github.com/jandedobbeleer/oh-my-posh/src/segments/options"

type Haskell struct {
	Language

	StackGhc bool
}

const (
	StackGhcMode options.Option = "stack_ghc_mode"
)

func (h *Haskell) Template() string {
	return languageTemplate
}

func (h *Haskell) Enabled() bool {
	ghcRegex := `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`

	h.extensions = []string{"*.hs", "*.lhs", "stack.yaml", "package.yaml", "*.cabal", "cabal.project"}
	h.tooling = map[string]*cmd{
		"ghc": {
			executable: "ghc",
			args:       []string{"--numeric-version"},
			regex:      ghcRegex,
		},
		"stack": {
			executable: "stack",
			args:       []string{"ghc", "--", "--numeric-version"},
			regex:      ghcRegex,
		},
	}
	h.defaultTooling = []string{"ghc"}
	h.versionURLTemplate = "https://www.haskell.org/ghc/download_ghc_{{ .Major }}_{{ .Minor }}_{{ .Patch }}.html"

	switch h.options.String(StackGhcMode, "never") {
	case "always":
		h.defaultTooling = []string{"stack"}
		h.StackGhc = true
	case "package":
		_, err := h.env.HasParentFilePath("stack.yaml", false)
		if err == nil {
			h.defaultTooling = []string{"stack"}
			h.StackGhc = true
		}
	}

	return h.Language.Enabled()
}
