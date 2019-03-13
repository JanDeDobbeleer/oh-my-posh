package main

import "fmt"

type venv struct {
	props    *properties
	env      environmentInfo
	venvName string
}

const (
	//PythonIcon used to indicate a Python virtualenv is active
	PythonIcon Property = "python_icon"
)

func (v *venv) string() string {
	return fmt.Sprintf("%s %s", v.props.getString(PythonIcon, "PYTHON:"), v.venvName)
}

func (v *venv) init(props *properties, env environmentInfo) {
	v.props = props
	v.env = env
}

func (v *venv) enabled() bool {
	venvVars := []string{
		"VIRTUAL_ENV",
		"CONDA_ENV_PATH",
		"CONDA_DEFAULT_ENV",
	}
	var venv string
	for _, venvVar := range venvVars {
		venv = v.env.getenv(venvVar)
		if venv != "" {
			v.venvName = base(venv, v.env)
			return true
		}
	}
	return false
}
