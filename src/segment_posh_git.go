package main

import "strings"

type poshgit struct {
	props Properties
	env   Environment

	Status string
}

const (
	poshGitEnv = "POSH_GIT_STATUS"
)

func (p *poshgit) enabled() bool {
	status := p.env.getenv(poshGitEnv)
	p.Status = strings.TrimSpace(status)
	return p.Status != ""
}

func (p *poshgit) string() string {
	segmentTemplate := p.props.getString(SegmentTemplate, "{{ .Status }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  p,
		Env:      p.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (p *poshgit) init(props Properties, env Environment) {
	p.props = props
	p.env = env
}
