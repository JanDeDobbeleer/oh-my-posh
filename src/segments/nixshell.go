package segments

import (
	"path/filepath"
	"strings"
)

const (
	NONE = "none"
)

type NixShell struct {
	Base

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

// Hack to detect a `nix shell` (vs a `nix-shell`) by checking if PATH contains a nix store;
// a better way will be enabled by https://github.com/NixOS/nix/issues/6677.
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
