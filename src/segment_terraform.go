package main

type terraform struct {
	props         properties
	env           environmentInfo
	WorkspaceName string
}

func (tf *terraform) string() string {
	segmentTemplate := tf.props.getString(SegmentTemplate, "{{.WorkspaceName}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  tf,
		Env:      tf.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (tf *terraform) init(props properties, env environmentInfo) {
	tf.props = props
	tf.env = env
}

func (tf *terraform) enabled() bool {
	cmd := "terraform"
	if !tf.env.hasCommand(cmd) || !tf.env.hasFolder(tf.env.getcwd()+"/.terraform") {
		return false
	}
	tf.WorkspaceName, _ = tf.env.runCommand(cmd, "workspace", "show")
	return true
}
