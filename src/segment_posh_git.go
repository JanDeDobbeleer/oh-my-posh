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

func (p *poshgit) template() string {
	return "{{ .Status }}"
}

func (p *poshgit) enabled() bool {
	status := p.env.Getenv(poshGitEnv)
	p.Status = strings.TrimSpace(status)
	return p.Status != ""
}

func (p *poshgit) init(props Properties, env Environment) {
	p.props = props
	p.env = env
}
