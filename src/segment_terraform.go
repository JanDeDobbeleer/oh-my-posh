package main

import (
	"oh-my-posh/environment"
)

type terraform struct {
	props Properties
	env   environment.Environment

	WorkspaceName string
}

func (tf *terraform) template() string {
	return "{{ .WorkspaceName }}"
}

func (tf *terraform) init(props Properties, env environment.Environment) {
	tf.props = props
	tf.env = env
}

func (tf *terraform) enabled() bool {
	cmd := "terraform"
	if !tf.env.HasCommand(cmd) || !tf.env.HasFolder(tf.env.Pwd()+"/.terraform") {
		return false
	}
	tf.WorkspaceName, _ = tf.env.RunCommand(cmd, "workspace", "show")
	return true
}
