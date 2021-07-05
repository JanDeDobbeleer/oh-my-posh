package main

import "oh-my-posh/runtime"

type root struct {
	props *properties
	env   runtime.Environment
}

const (
	// RootIcon indicates the root user
	RootIcon Property = "root_icon"
)

func (rt *root) enabled() bool {
	return rt.env.IsRunningAsRoot()
}

func (rt *root) string() string {
	return rt.props.getString(RootIcon, "\uF0E7")
}

func (rt *root) init(props *properties, env runtime.Environment) {
	rt.props = props
	rt.env = env
}
