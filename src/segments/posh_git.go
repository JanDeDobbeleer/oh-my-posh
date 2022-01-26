package segments

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

func (p *PoshGit) Template() string {
	return "{{ .Status }}"
}

func (p *PoshGit) Enabled() bool {
	status := p.env.Getenv(poshGitEnv)
	p.Status = strings.TrimSpace(status)
	return p.Status != ""
}

func (p *PoshGit) Init(props properties.Properties, env environment.Environment) {
	p.props = props
	p.env = env
}
