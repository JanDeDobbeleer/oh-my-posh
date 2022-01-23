package main

type ruby struct {
	language
}

func (r *ruby) template() string {
	return languageTemplate
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
	enabled := r.language.enabled()
	// this happens when no version is set
	if r.Full == "______" {
		r.Full = ""
	}
	return enabled
}
