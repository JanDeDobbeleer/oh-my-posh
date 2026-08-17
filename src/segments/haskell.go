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
	h.loadSpec()

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

// Activation implements the activation gate; see Language.activation.
func (h *Haskell) Activation() Activation {
	h.loadSpec()

	return h.activation()
}

func (h *Haskell) loadSpec() {
	h.extensions = []string{"*.hs", "*.lhs", "stack.yaml", "package.yaml", "*.cabal", "cabal.project"}
	h.tooling = map[string]*cmd{
		ghcToolName: {
			executable:       ghcToolName,
			args:             []string{"--numeric-version"},
			regex:            versionRegex,
			versionCacheable: true,
		},
		// Not marked versionCacheable: stack resolves the GHC version to run
		// from the project's stack.yaml resolver, so the same stack binary
		// reports a different version per project.
		stackToolName: {
			executable: stackToolName,
			args:       []string{ghcToolName, "--", "--numeric-version"},
			regex:      versionRegex,
		},
	}
	h.defaultTooling = []string{ghcToolName}
	h.versionURLTemplate = "https://www.haskell.org/ghc/download_ghc_{{ .Major }}_{{ .Minor }}_{{ .Patch }}.html"
}
