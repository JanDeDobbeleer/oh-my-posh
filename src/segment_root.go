package main

type root struct {
	props Properties
	env   environmentInfo
}

const (
	// RootIcon indicates the root user
	RootIcon Property = "root_icon"
)

func (rt *root) enabled() bool {
	return rt.env.isRunningAsRoot()
}

func (rt *root) string() string {
	return rt.props.getString(RootIcon, "\uF0E7")
}

func (rt *root) init(props Properties, env environmentInfo) {
	rt.props = props
	rt.env = env
}
