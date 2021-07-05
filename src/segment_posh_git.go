package main

import (
	"strings"

	"oh-my-posh/runtime"
)

type poshgit struct {
	props     *properties
	env       runtime.Environment
	gitStatus string
}

const (
	poshGitEnv = "POSH_GIT_STATUS"
)

func (p *poshgit) enabled() bool {
	status := p.env.Getenv(poshGitEnv)
	p.gitStatus = strings.TrimSpace(status)
	return p.gitStatus != ""
}

func (p *poshgit) string() string {
	return p.gitStatus
}

func (p *poshgit) init(props *properties, env runtime.Environment) {
	p.props = props
	p.env = env
}
