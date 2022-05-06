package segments

import (
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type ITerm struct {
	props properties.Properties
	env   environment.Environment

	PromptMark string
}

func (i *ITerm) Template() string {
	return "{{ .PromptMark }}"
}

func (i *ITerm) Enabled() bool {
	promptMark, err := i.getResult()
	if err != nil {
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

	// Check to ensure the user has squelched the default mark
	if i.env.Getenv("ITERM2_SQUELCH_MARK") != "1" {
		return "", errors.New("iTerm default mark enabled (set ITERM2_SQUELCH_MARK=1)")
	}

	// Now, set the mark string based on shell (or error out)
	switch i.env.Shell() {
	case "zsh":
		response = `%{$(iterm2_prompt_mark)%}`
	case "bash":
		response = `\[$(iterm2_prompt_mark)\]`
	case "fish":
		response = `iterm2_prompt_mark`
	default:
		return "", errors.New("Shell isn't compatible with iTerm Shell Integration")
	}

	return response, nil
}

func (i *ITerm) Init(props properties.Properties, env environment.Environment) {
	i.props = props
	i.env = env
}
