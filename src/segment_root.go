package main

type root struct {
	props Properties
	env   Environment
}

func (rt *root) template() string {
	return "\uF0E7"
}

func (rt *root) enabled() bool {
	return rt.env.Root()
}

func (rt *root) init(props Properties, env Environment) {
	rt.props = props
	rt.env = env
}
