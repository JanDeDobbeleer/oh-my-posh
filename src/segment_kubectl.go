package main

import (
	"strings"
)

type kubectl struct {
	props     *properties
	env       environmentInfo
	Context   string
	Namespace string
}

func (k *kubectl) string() string {
	segmentTemplate := k.props.getString(SegmentTemplate, "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  k,
		Env:      k.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (k *kubectl) init(props *properties, env environmentInfo) {
	k.props = props
	k.env = env
}

func (k *kubectl) enabled() bool {
	cmd := "kubectl"
	if !k.env.hasCommand(cmd) {
		return false
	}
	result, err := k.env.runCommand(cmd, "config", "view", "--minify", "--output", "jsonpath={..current-context},{..namespace}")
	displayError := k.props.getBool(DisplayError, false)
	if err != nil && displayError {
		k.Context = "KUBECTL ERR"
		k.Namespace = k.Context
		return true
	}
	if err != nil {
		return false
	}

	values := strings.Split(result, ",")
	k.Context = values[0]
	k.Namespace = values[1]
	return k.Context != ""
}
