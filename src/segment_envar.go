package main

import "oh-my-posh/runtime"

type envvar struct {
	props   *properties
	env     runtime.Environment
	content string
}

const (
	// VarName name of the variable
	VarName Property = "var_name"
)

func (e *envvar) enabled() bool {
	name := e.props.getString(VarName, "")
	e.content = e.env.Getenv(name)
	return e.content != ""
}

func (e *envvar) string() string {
	return e.content
}

func (e *envvar) init(props *properties, env runtime.Environment) {
	e.props = props
	e.env = env
}
