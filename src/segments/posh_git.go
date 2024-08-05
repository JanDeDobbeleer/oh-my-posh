package segments

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	poshGitEnv = "POSH_GIT_STATUS"
)

type poshGit struct {
	Index        *poshGitStatus `json:"Index"`
	Working      *poshGitStatus `json:"Working"`
	RepoName     string         `json:"RepoName"`
	Branch       string         `json:"Branch"`
	GitDir       string         `json:"GitDir"`
	Upstream     string         `json:"Upstream"`
	StashCount   int            `json:"StashCount"`
	AheadBy      int            `json:"AheadBy"`
	BehindBy     int            `json:"BehindBy"`
	HasWorking   bool           `json:"HasWorking"`
	HasIndex     bool           `json:"HasIndex"`
	HasUntracked bool           `json:"HasUntracked"`
}

type poshGitStatus struct {
	Added    []string `json:"Added"`
	Modified []string `json:"Modified"`
	Deleted  []string `json:"Deleted"`
	Unmerged []string `json:"Unmerged"`
}

func (s *GitStatus) parsePoshGitStatus(p *poshGitStatus) {
	if p == nil {
		return
	}

	s.Added = len(p.Added)
	s.Deleted = len(p.Deleted)
	s.Modified = len(p.Modified)
	s.Unmerged = len(p.Unmerged)
}

func (g *Git) hasPoshGitStatus() bool {
	envStatus := g.env.Getenv(poshGitEnv)
	if len(envStatus) == 0 {
		g.env.Error(fmt.Errorf("%s environment variable not set, do you have the posh-git module installed?", poshGitEnv))
		return false
	}

	var posh poshGit
	err := json.Unmarshal([]byte(envStatus), &posh)
	if err != nil {
		g.env.Error(err)
		return false
	}

	g.setDir(posh.GitDir)
	g.Working = &GitStatus{}
	g.Working.parsePoshGitStatus(posh.Working)
	g.Staging = &GitStatus{}
	g.Staging.parsePoshGitStatus(posh.Index)
	g.HEAD = g.parsePoshGitHEAD(posh.Branch)
	g.stashCount = posh.StashCount
	g.Ahead = posh.AheadBy
	g.Behind = posh.BehindBy
	g.UpstreamGone = len(posh.Upstream) == 0
	g.Upstream = posh.Upstream

	g.setBranchStatus()

	if len(g.Upstream) != 0 && g.props.GetBool(FetchUpstreamIcon, false) {
		g.UpstreamIcon = g.getUpstreamIcon()
	}

	g.poshgit = true
	return true
}

func (g *Git) parsePoshGitHEAD(head string) string {
	// commit
	if strings.HasSuffix(head, "...)") {
		head = strings.TrimLeft(head, "(")
		head = strings.TrimRight(head, ".)")
		return fmt.Sprintf("%s%s", g.props.GetString(CommitIcon, "\uF417"), head)
	}
	// tag
	if strings.HasPrefix(head, "(") {
		head = strings.TrimLeft(head, "(")
		head = strings.TrimRight(head, ")")
		return fmt.Sprintf("%s%s", g.props.GetString(TagIcon, "\uF412"), head)
	}
	// regular branch
	return fmt.Sprintf("%s%s", g.props.GetString(BranchIcon, "\uE0A0"), g.formatBranch(head))
}
