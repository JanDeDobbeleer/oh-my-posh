package segments

import (
	"path/filepath"
	"strings"
)

const (
	NONE = "none"
)

type NixShell struct {
	base

	Type string
}

func (n *NixShell) Template() string {
	return "via {{ .Type }}-shell"
}

func (n *NixShell) DetectType() string {
	shellType := n.env.Getenv("IN_NIX_SHELL")

	switch shellType {
	case "pure", "impure":
		return shellType
	default:
		if n.InNewNixShell() {
			return UNKNOWN
		}

		return NONE
	}
}

// Hack to detect if we're in a `nix shell` (in contrast to a `nix-shell`).
// A better way to do this will be enabled by https://github.com/NixOS/nix/issues/6677
// so we check if the PATH contains a nix store.
func (n *NixShell) InNewNixShell() bool {
	paths := filepath.SplitList(n.env.Getenv("PATH"))

	for _, p := range paths {
		if strings.Contains(p, "/nix/store") {
			return true
		}
	}

	return false
}

func (n *NixShell) Enabled() bool {
	n.Type = n.DetectType()

	return n.Type != NONE
}
