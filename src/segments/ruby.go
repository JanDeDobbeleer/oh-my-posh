package segments

type Ruby struct {
	Language
}

func (r *Ruby) Template() string {
	return languageTemplate
}

func (r *Ruby) Enabled() bool {
	r.extensions = []string{"*.rb", "Rakefile", "Gemfile"}
	r.tooling = map[string]*cmd{
		"rbenv": {
			executable: "rbenv",
			args:       []string{"version-name"},
			regex:      `(?P<version>.+)`,
		},
		"rvm-prompt": {
			executable: "rvm-prompt",
			args:       []string{"i", "v", "g"},
			regex:      `(?P<version>.+)`,
		},
		"chruby": {
			executable: "chruby",
			args:       []string(nil),
			regex:      `\* (?P<version>.+)\n`,
		},
		"asdf": {
			executable: "asdf",
			args:       []string{"current", "ruby"},
			regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
		},
		"ruby": {
			executable: "ruby",
			args:       []string{"--version"},
			regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
		},
	}
	r.defaultTooling = []string{"rbenv", "rvm-prompt", "chruby", "asdf", "ruby"}

	enabled := r.Language.Enabled()

	// this happens when no version is set
	if r.Full == "______" {
		r.Full = ""
	}

	return enabled
}
