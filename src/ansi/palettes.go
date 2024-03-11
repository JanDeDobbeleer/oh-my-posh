package ansi

type Palettes struct {
	Template string             `json:"template,omitempty" toml:"template,omitempty"`
	List     map[string]Palette `json:"list,omitempty" toml:"list,omitempty"`
}
