package segments

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

type ITerm struct {
	props properties.Properties
	env   platform.Environment

	PromptMark string
}

func (i *ITerm) Template() string {
	return "{{ .PromptMark }}"
}

func (i *ITerm) Enabled() bool {
	promptMark, err := i.getResult()
	if err != nil {
		i.env.Error(err)
		return false
	}
	i.PromptMark = promptMark

	return true
}

func (i *ITerm) getResult() (string, error) {
	var response string
	// First, check if we're using iTerm
	if i.env.Getenv("TERM_PROGRAM") != "iTerm.app" {
		return "", errors.New("Only works with iTerm")
	}

	// Check to ensure the user has squelched the default mark for BASH and ZSH
	if i.env.Getenv("ITERM2_SQUELCH_MARK") != "1" {
		return "", errors.New("iTerm default mark enabled (export ITERM2_SQUELCH_MARK=1)")
	}

	// Now, set the mark string based on shell (or error out)
	switch i.env.Shell() {
	case shell.ZSH:
		response = `%{$(iterm2_prompt_mark)%}`
	case shell.BASH:
		response = `\[$(iterm2_prompt_mark)\]`
	default:
		return "", errors.New("Shell isn't compatible with iTerm Shell Integration")
	}

	return response, nil
}

func (i *ITerm) Init(props properties.Properties, env platform.Environment) {
	i.props = props
	i.env = env
}
