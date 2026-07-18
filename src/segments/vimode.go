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
	case "main", "viins", "emacs", "insert":
		return "insert"
	case "vicmd", "default":
		return "normal"
	case "visual":
		return "visual"
	case "viopp", "operator", "f", "F", "t", "T":
		return "viopp"
	case "replace", "replace_one":
		return "replace"
	default:
		return keymap
	}
}
