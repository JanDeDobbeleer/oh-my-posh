package segments

type Ruby struct {
	language
}

func (r *Ruby) Template() string {
	return languageTemplate
}

func (r *Ruby) Enabled() bool {
	r.extensions = []string{"*.rb", "Rakefile", "Gemfile"}
	r.commands = []*cmd{
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
	}

	enabled := r.language.Enabled()

	// this happens when no version is set
	if r.Full == "______" {
		r.Full = ""
	}

	return enabled
}
