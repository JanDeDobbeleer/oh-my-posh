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
	}
	return template.render()
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
	if err != nil {
		k.Context = "KUBECTL ERR"
		k.Namespace = k.Context
		return true
	}

	values := strings.Split(result, ",")
	k.Context = values[0]
	k.Namespace = values[1]
	return k.Context != ""
}
