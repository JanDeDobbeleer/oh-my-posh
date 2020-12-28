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
	commandPath, commandExists := k.env.hasCommand("kubectl")
	if !commandExists {
		return false
	}
	k.contextName, _ = k.env.runCommand(commandPath, "config", "current-context")
	return k.contextName != ""
}
