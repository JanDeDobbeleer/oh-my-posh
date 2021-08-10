package main

type text struct {
	props *properties
	env   environmentInfo
}

const (
	// TextProperty represents text to write
	TextProperty Property = "text"
)

func (t *text) enabled() bool {
	return true
}

func (t *text) string() string {
	textProperty := t.props.getString(TextProperty, "!!text property not defined!!")
	template := &textTemplate{
		Template: textProperty,
		Context:  t,
		Env:      t.env,
	}
	textOutput := template.renderPlainContextTemplate(nil)
	return textOutput
}

func (t *text) init(props *properties, env environmentInfo) {
	t.props = props
	t.env = env
}
