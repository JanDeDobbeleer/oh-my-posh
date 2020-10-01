package main

type node struct {
	props       *properties
	env         environmentInfo
	nodeVersion string
}

func (n *node) string() string {
	return n.nodeVersion
}

func (n *node) init(props *properties, env environmentInfo) {
	n.props = props
	n.env = env
}

func (n *node) enabled() bool {
	if !n.env.hasFiles("*.js") && !n.env.hasFiles("*.ts") {
		return false
	}
	if !n.env.hasCommand("node") {
		return false
	}
	n.nodeVersion = n.env.runCommand("node", "--version")
	return true
}
