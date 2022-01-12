package main

type text struct {
	props   Properties
	env     Environment
	content string
}

const (
	// TextProperty represents text to write
	TextProperty Property = "text"
)

func (t *text) enabled() bool {
	textProperty := t.props.getOneOfString(TextProperty, SegmentTemplate, "!!text property not defined!!")
	template := &textTemplate{
		Template: textProperty,
		Context:  t,
		Env:      t.env,
	}
	if text, err := template.render(); err == nil {
		t.content = text
		return len(t.content) > 0
	}
	return false
}

func (t *text) string() string {
	return t.content
}

func (t *text) init(props Properties, env Environment) {
	t.props = props
	t.env = env
}
