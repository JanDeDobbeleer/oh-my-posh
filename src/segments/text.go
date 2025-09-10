package segments

type Text struct {
	Dummy struct{}
	base
}

func (t *Text) Template() string {
	return "  "
}

func (t *Text) Enabled() bool {
	return true
}
