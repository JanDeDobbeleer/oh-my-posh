package main

type nvm struct {
	props       *properties
	env         environmentInfo
	nodeVersion string
}

func (n *nvm) string() string {
	return n.nodeVersion
}

func (n *nvm) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}

func (n *nvm) enabled() bool {
	if !n.env.hasCommand("node") {
		return false
	}
	n.nodeVersion = n.env.runCommand("node", "--version")
	return true
}
