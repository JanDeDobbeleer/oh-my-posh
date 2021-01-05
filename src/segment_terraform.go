package main

type terraform struct {
	props         *properties
	env           environmentInfo
	workspaceName string
}

func (tf *terraform) string() string {
	return tf.workspaceName
}

func (tf *terraform) init(props *properties, env environmentInfo) {
	tf.props = props
	tf.env = env
}

func (tf *terraform) enabled() bool {
	cmd := "terraform"
	if !tf.env.hasCommand(cmd) || !tf.env.hasFolder(".terraform") {
		return false
	}
	tf.workspaceName, _ = tf.env.runCommand(cmd, "workspace", "show")
	return true
}
