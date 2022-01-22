package main

type rust struct {
	language
}

func (r *rust) string() string {
	segmentTemplate := r.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return r.language.string()
	}
	return r.language.renderTemplate(segmentTemplate, r)
}

func (r *rust) init(props Properties, env Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		commands: []*cmd{
			{
				executable: "rustc",
				args:       []string{"--version"},
				regex:      `rustc (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
	}
}

func (r *rust) enabled() bool {
	return r.language.enabled()
}
