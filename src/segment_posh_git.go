package main

import (
	"oh-my-posh/environment"
	"strings"
)

type poshgit struct {
	props Properties
	env   environment.Environment

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

func (p *poshgit) init(props Properties, env environment.Environment) {
	p.props = props
	p.env = env
}
