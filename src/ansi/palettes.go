package ansi

type Palettes struct {
	Template string             `json:"template,omitempty"`
	List     map[string]Palette `json:"list,omitempty"`
}
