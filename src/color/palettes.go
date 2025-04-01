package color

type Palettes struct {
	List     map[string]Palette `json:"list,omitempty" toml:"list,omitempty" yaml:"list,omitempty"`
	Template string             `json:"template,omitempty" toml:"template,omitempty" yaml:"template,omitempty"`
}
