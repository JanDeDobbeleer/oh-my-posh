package main

type kubectl struct {
	props       *properties
	env         environmentInfo
	contextName string
}

func (k *kubectl) string() string {
	return k.contextName
}

func (k *kubectl) init(props *properties, env environmentInfo) {
	k.props = props
	k.env = env
}

func (k *kubectl) enabled() bool {
	if !k.env.hasCommand("kubectl") {
		return false
	}
	k.contextName, _ = k.env.runCommand("kubectl", "config", "current-context")
	return k.contextName != ""
}
