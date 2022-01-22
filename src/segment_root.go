package main

type root struct {
	props Properties
	env   Environment
}

func (rt *root) enabled() bool {
	return rt.env.isRunningAsRoot()
}

func (rt *root) string() string {
	segmentTemplate := rt.props.getString(SegmentTemplate, "\uF0E7")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  rt,
		Env:      rt.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (rt *root) init(props Properties, env Environment) {
	rt.props = props
	rt.env = env
}
