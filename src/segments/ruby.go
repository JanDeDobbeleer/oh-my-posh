package segments

type Ruby struct {
	Language
}

func (r *Ruby) Template() string {
	return languageTemplate
}

const (
	chrubyToolName    = "chruby"
	rbenvToolName     = "rbenv"
	rvmPromptToolName = "rvm-prompt"
	rubyToolName      = "ruby"
)

func (r *Ruby) Enabled() bool {
	r.extensions = []string{"*.rb", "Rakefile", "Gemfile"}
	r.tooling = map[string]*cmd{
		rbenvToolName: {
			executable: rbenvToolName,
			args:       []string{"version-name"},
			regex:      `(?P<version>.+)`,
		},
		rvmPromptToolName: {
			executable: rvmPromptToolName,
			args:       []string{"i", "v", "g"},
			regex:      `(?P<version>.+)`,
		},
		chrubyToolName: {
			executable: chrubyToolName,
			args:       []string(nil),
			regex:      `\* (?P<version>.+)\n`,
		},
		asdfToolName: {
			executable: asdfToolName,
			args:       []string{"current", rubyToolName},
			regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
		},
		rubyToolName: {
			executable: rubyToolName,
			args:       []string{versionFlagArg},
			regex:      `ruby\s+(?P<version>[^\s]+)\s+`,
		},
	}
	r.defaultTooling = []string{rbenvToolName, rvmPromptToolName, chrubyToolName, asdfToolName, rubyToolName}

	enabled := r.Language.Enabled()

	// this happens when no version is set
	if r.Full == "______" {
		r.Full = ""
	}

	return enabled
}
