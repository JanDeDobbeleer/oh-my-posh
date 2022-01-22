package main

type ruby struct {
	language
}

func (r *ruby) string() string {
	segmentTemplate := r.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		version := r.language.string()
		// asdf default non-set version
		if version == "______" {
			return ""
		}
		return version
	}
	return r.language.renderTemplate(segmentTemplate, r)
}

func (r *ruby) init(props Properties, env Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.rb", "Rakefile", "Gemfile"},
		commands: []*cmd{
			{
				executable: "rbenv",
				args:       []string{"version-name"},
				regex:      `(?P<version>.+)`,
			},
			{
				executable: "rvm-prompt",
				args:       []string{"i", "v", "g"},
				regex:      `(?P<version>.+)`,
			},
			{
				executable: "chruby",
				args:       []string(nil),
				regex:      `\* (?P<version>.+)\n`,
			},
			{
				executable: "asdf",
				args:       []string{"current", "ruby"},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
			{
				executable: "ruby",
				args:       []string{"--version"},
				regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
			},
		},
	}
}

func (r *ruby) enabled() bool {
	return r.language.enabled()
}
