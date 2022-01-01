package main

type text struct {
	props   Properties
	env     environmentInfo
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
	t.content = template.renderPlainContextTemplate(nil)
	return len(t.content) > 0
}

func (t *text) string() string {
	return t.content
}

func (t *text) init(props Properties, env environmentInfo) {
	t.props = props
	t.env = env
}
