package main

import "strings"

type poshgit struct {
	props     properties
	env       environmentInfo
	gitStatus string
}

const (
	poshGitEnv = "POSH_GIT_STATUS"
)

func (p *poshgit) enabled() bool {
	status := p.env.getenv(poshGitEnv)
	p.gitStatus = strings.TrimSpace(status)
	return p.gitStatus != ""
}

func (p *poshgit) string() string {
	return p.gitStatus
}

func (p *poshgit) init(props properties, env environmentInfo) {
	p.props = props
	p.env = env
}
