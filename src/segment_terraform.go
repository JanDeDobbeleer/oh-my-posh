package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Terraform struct {
	props properties.Properties
	env   environment.Environment

	WorkspaceName string
}

func (tf *Terraform) Template() string {
	return "{{ .WorkspaceName }}"
}

func (tf *Terraform) Init(props properties.Properties, env environment.Environment) {
	tf.props = props
	tf.env = env
}

func (tf *Terraform) Enabled() bool {
	cmd := "terraform"
	if !tf.env.HasCommand(cmd) || !tf.env.HasFolder(tf.env.Pwd()+"/.terraform") {
		return false
	}
	tf.WorkspaceName, _ = tf.env.RunCommand(cmd, "workspace", "show")
	return true
}
