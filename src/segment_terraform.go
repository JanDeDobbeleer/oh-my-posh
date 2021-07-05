package main

import "oh-my-posh/runtime"

type terraform struct {
	props         *properties
	env           runtime.Environment
	workspaceName string
}

func (tf *terraform) string() string {
	return tf.workspaceName
}

func (tf *terraform) init(props *properties, env runtime.Environment) {
	tf.props = props
	tf.env = env
}

func (tf *terraform) enabled() bool {
	cmd := "terraform"
	if !tf.env.HasCommand(cmd) || !tf.env.HasFolder(".terraform") {
		return false
	}
	tf.workspaceName, _ = tf.env.RunCommand(cmd, "workspace", "show")
	return true
}
