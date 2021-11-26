package main

type envvar struct {
	props properties
	env   environmentInfo
	Value string
}

const (
	// VarName name of the variable
	VarName Property = "var_name"
)

func (e *envvar) enabled() bool {
	name := e.props.getString(VarName, "")
	e.Value = e.env.getenv(name)
	return e.Value != ""
}

func (e *envvar) string() string {
	segmentTemplate := e.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return e.Value
	}
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  e,
		Env:      e.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (e *envvar) init(props properties, env environmentInfo) {
	e.props = props
	e.env = env
}
