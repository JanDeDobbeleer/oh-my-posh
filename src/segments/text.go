package segments

type Text struct {
	base
	Dummy int
}

func (t *Text) Template() string {
	return "  "
}

func (t *Text) Enabled() bool {
	return true
}
