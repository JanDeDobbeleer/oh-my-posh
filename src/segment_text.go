package main

type text struct {
	props Properties
	env   Environment

	Text string
}

const (
	// TextProperty represents text to write
	TextProperty Property = "text"
)

func (t *text) enabled() bool {
	segmentTemplate := t.props.getString(SegmentTemplate, "{{ .Text }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  t,
		Env:      t.env,
	}
	if text, err := template.render(); err == nil {
		t.Text = text
		return len(t.Text) > 0
	}
	return false
}

func (t *text) string() string {
	return t.Text
}

func (t *text) init(props Properties, env Environment) {
	t.props = props
	t.env = env
}
