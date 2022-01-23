package main

type text struct {
	props Properties
	env   Environment

	Text string
}

func (t *text) template() string {
	return "{{ .Text }}"
}

func (t *text) enabled() bool {
	return true
}

func (t *text) init(props Properties, env Environment) {
	t.props = props
	t.env = env
}
