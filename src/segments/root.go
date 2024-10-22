package segments

type Root struct {
	base
}

func (rt *Root) Template() string {
	return " \uF0E7 "
}

func (rt *Root) Enabled() bool {
	return rt.env.Root()
}
