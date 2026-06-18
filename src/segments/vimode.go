package segments

const (
	poshVIModeEnv = "POSH_VI_MODE"
)

type VIMode struct {
	Base

	Mode   string
	Keymap string
}

func (v *VIMode) Template() string {
	return " {{ .Mode }} "
}

func (v *VIMode) Enabled() bool {
	v.Keymap = v.env.Getenv(poshVIModeEnv)
	if v.Keymap == "" {
		return false
	}

	v.Mode = mapVIModeKeymap(v.Keymap)
	return true
}

func mapVIModeKeymap(keymap string) string {
	switch keymap {
	case "main", "viins", "emacs":
		return "insert"
	case "vicmd":
		return "normal"
	case "visual":
		return "visual"
	case "viopp":
		return "viopp"
	case "replace":
		return "replace"
	default:
		return keymap
	}
}
