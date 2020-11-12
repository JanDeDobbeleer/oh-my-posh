package main

type envvar struct {
	props   *properties
	env     environmentInfo
	content string
}

const (
	// VarName name of the variable
	VarName Property = "var_name"
)

func (e *envvar) enabled() bool {
	name := e.props.getString(VarName, "")
	e.content = e.env.getenv(name)
	return e.content != ""
}

func (e *envvar) string() string {
	return e.content
}

func (e *envvar) init(props *properties, env environmentInfo) {
	e.props = props
	e.env = env
}
