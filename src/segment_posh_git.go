package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"strings"
)

type PoshGit struct {
	props properties.Properties
	env   environment.Environment

	Status string
}

const (
	poshGitEnv = "POSH_GIT_STATUS"
)

func (p *PoshGit) template() string {
	return "{{ .Status }}"
}

func (p *PoshGit) enabled() bool {
	status := p.env.Getenv(poshGitEnv)
	p.Status = strings.TrimSpace(status)
	return p.Status != ""
}

func (p *PoshGit) init(props properties.Properties, env environment.Environment) {
	p.props = props
	p.env = env
}
