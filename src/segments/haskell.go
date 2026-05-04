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

const ghcToolName = "ghc"

const stackToolName = "stack"

func (h *Haskell) Enabled() bool {
	h.extensions = []string{"*.hs", "*.lhs", "stack.yaml", "package.yaml", "*.cabal", "cabal.project"}
	h.tooling = map[string]*cmd{
		ghcToolName: {
			executable: ghcToolName,
			args:       []string{"--numeric-version"},
			regex:      versionRegex,
		},
		stackToolName: {
			executable: stackToolName,
			args:       []string{ghcToolName, "--", "--numeric-version"},
			regex:      versionRegex,
		},
	}
	h.defaultTooling = []string{ghcToolName}
	h.versionURLTemplate = "https://www.haskell.org/ghc/download_ghc_{{ .Major }}_{{ .Minor }}_{{ .Patch }}.html"

	switch h.options.String(StackGhcMode, "never") {
	case "always":
		h.defaultTooling = []string{stackToolName}
		h.StackGhc = true
	case "package":
		_, err := h.env.HasParentFilePath("stack.yaml", false)
		if err == nil {
			h.defaultTooling = []string{stackToolName}
			h.StackGhc = true
		}
	}

	return h.Language.Enabled()
}
