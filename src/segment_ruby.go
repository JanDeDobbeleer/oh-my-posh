package main

type ruby struct {
	props   *properties
	env     environmentInfo
	version string
}

func (r *ruby) string() string {
	if r.props.getBool(DisplayVersion, true) {
		return r.version
	}
	return ""
}

func (r *ruby) init(props *properties, env environmentInfo) {
	r.props = props
	r.env = env
}

func (r *ruby) enabled() bool {
	if !r.env.hasFiles("*.rb") && !r.env.hasFiles("Rakefile") && !r.env.hasFiles("Gemfile") {
		return false
	}
	if !r.props.getBool(DisplayVersion, true) {
		return true
	}
	r.version = r.getVersion()
	return r.version != ""
}

func (r *ruby) getVersion() string {
	options := []struct {
		Command string
		Args    []string
		Regex   string
	}{
		{Command: "rbenv", Args: []string{"version-name"}, Regex: `(?P<version>.+)`},
		{Command: "rvm-prompt", Args: []string{"i", "v", "g"}, Regex: `(?P<version>.+)`},
		{Command: "chruby", Args: []string(nil), Regex: `\* (?P<version>.+)\n`},
		{Command: "asdf", Args: []string{"current", "ruby"}, Regex: `ruby\s+(?P<version>[^\s_]+)\s+`},
		{Command: "ruby", Args: []string{"--version"}, Regex: `ruby\s+(?P<version>[^\s_]+)\s+`},
	}
	for _, option := range options {
		if !r.env.hasCommand(option.Command) {
			continue
		}
		version, _ := r.env.runCommand(option.Command, option.Args...)
		match := findNamedRegexMatch(option.Regex, version)
		if match["version"] == "" {
			continue
		}
		return match["version"]
	}
	return ""
}
