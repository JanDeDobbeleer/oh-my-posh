package segments

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

type ITerm struct {
	props properties.Properties
	env   platform.Environment
}

func (i *ITerm) Template() string {
	return "{{ .PromptMark }}"
}

func (i *ITerm) Enabled() bool {
	return i.env.Getenv("TERM_PROGRAM") == "iTerm.app"
}

func (i *ITerm) PromptMark() string {
	// Check to ensure the user has squelched the default mark for BASH and ZSH
	if i.env.Getenv("ITERM2_SQUELCH_MARK") != "1" {
		i.env.Debug("iTerm default mark enabled, adjust using export ITERM2_SQUELCH_MARK=1")
		return ""
	}

	sh := i.env.Shell()
	if sh != shell.ZSH && sh != shell.BASH {
		i.env.Debug("Shell is not ZSH or BASH, cannot set prompt mark")
		return ""
	}

	return i.format("$(iterm2_prompt_mark)")
}

func (i *ITerm) CurrentDir() string {
	dir := fmt.Sprintf("\x1b]1337;CurrentDir=%s\x07", i.env.Pwd())
	return i.format(dir)
}

func (i *ITerm) RemoteHost() string {
	host, err := i.env.Host()
	if err != nil {
		return ""
	}

	remoteHost := fmt.Sprintf("\x1b]1337;RemoteHost=%s@%s\x07", i.env.User(), host)
	return i.format(remoteHost)
}

func (i *ITerm) format(input string) string {
	switch i.env.Shell() {
	case shell.ZSH:
		return fmt.Sprintf(`%%{%s%%}`, input)
	case shell.BASH:
		return fmt.Sprintf(`\[%s\]`, input)
	default:
		return input
	}
}

func (i *ITerm) Init(props properties.Properties, env platform.Environment) {
	i.props = props
	i.env = env
}
