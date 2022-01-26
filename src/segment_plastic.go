package main

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/regex"
	"strconv"
	"strings"
)

type PlasticStatus struct {
	ScmStatus
}

func (s *PlasticStatus) add(code string) {
	switch code {
	case "LD":
		s.Deleted++
	case "AD", "PR":
		s.Added++
	case "LM":
		s.Moved++
	case "CH", "CO":
		s.Modified++
	}
}

type plastic struct {
	scm

	Status       *PlasticStatus
	Behind       bool
	Selector     string
	MergePending bool

	plasticWorkspaceFolder string // root folder of workspace
}

func (p *plastic) init(props Properties, env environment.Environment) {
	p.props = props
	p.env = env
}

func (p *plastic) template() string {
	return "{{ .Selector }}"
}

func (p *plastic) enabled() bool {
	if !p.env.HasCommand("cm") {
		return false
	}
	wkdir, err := p.env.HasParentFilePath(".plastic")
	if err != nil {
		return false
	}
	if p.shouldIgnoreRootRepository(wkdir.ParentFolder) {
		return false
	}
	if !wkdir.IsDir {
		return false
	}
	p.plasticWorkspaceFolder = wkdir.ParentFolder
	displayStatus := p.props.GetBool(FetchStatus, false)
	p.setSelector()
	if displayStatus {
		p.setPlasticStatus()
	}
	return true
}

func (p *plastic) setPlasticStatus() {
	output := p.getCmCommandOutput("status", "--all", "--machinereadable")
	splittedOutput := strings.Split(output, "\n")
	// compare to head
	currentChangeset := p.parseStatusChangeset(splittedOutput[0])
	headChangeset := p.getHeadChangeset()
	p.Behind = headChangeset > currentChangeset

	// parse file state
	p.MergePending = false
	p.Status = &PlasticStatus{}
	p.parseFilesStatus(splittedOutput)
}

func (p *plastic) parseFilesStatus(output []string) {
	if len(output) <= 1 {
		return
	}
	for _, line := range output[1:] {
		if len(line) < 3 {
			continue
		}

		if strings.Contains(line, "NO_MERGES") {
			p.Status.Unmerged++
			continue
		}

		p.MergePending = p.MergePending || regex.MatchString(`(?i)\smerge\s+from\s+[0-9]+\s*$`, line)

		code := line[:2]
		p.Status.add(code)
	}
}

func (p *plastic) parseStringPattern(output, pattern, name string) string {
	match := regex.FindNamedRegexMatch(pattern, output)
	if sValue, ok := match[name]; ok {
		return sValue
	}
	return ""
}

func (p *plastic) parseIntPattern(output, pattern, name string, defValue int) int {
	sValue := p.parseStringPattern(output, pattern, name)
	if len(sValue) > 0 {
		iValue, _ := strconv.Atoi(sValue)
		return iValue
	}
	return defValue
}

func (p *plastic) parseStatusChangeset(status string) int {
	return p.parseIntPattern(status, `STATUS\s+(?P<cs>[0-9]+?)\s`, "cs", 0)
}

func (p *plastic) getHeadChangeset() int {
	output := p.getCmCommandOutput("status", "--head", "--machinereadable")
	return p.parseIntPattern(output, `\bcs:(?P<cs>[0-9]+?)\s`, "cs", 0)
}

func (p *plastic) setSelector() {
	var ref string
	selector := p.FileContents(p.plasticWorkspaceFolder+"/.plastic/", "plastic.selector")
	// changeset
	ref = p.parseChangesetSelector(selector)
	if len(ref) > 0 {
		p.Selector = fmt.Sprintf("%s%s", p.props.GetString(CommitIcon, "\uF417"), ref)
		return
	}
	// fallback to label
	ref = p.parseLabelSelector(selector)
	if len(ref) > 0 {
		p.Selector = fmt.Sprintf("%s%s", p.props.GetString(TagIcon, "\uF412"), ref)
		return
	}
	// fallback to branch/smartbranch
	ref = p.parseBranchSelector(selector)
	if len(ref) > 0 {
		ref = p.truncateBranch(ref)
	}
	p.Selector = fmt.Sprintf("%s%s", p.props.GetString(BranchIcon, "\uE0A0"), ref)
}

func (p *plastic) parseChangesetSelector(selector string) string {
	return p.parseStringPattern(selector, `\bchangeset "(?P<cs>[0-9]+?)"`, "cs")
}

func (p *plastic) parseLabelSelector(selector string) string {
	return p.parseStringPattern(selector, `label "(?P<label>[a-zA-Z0-9\-\_]+?)"`, "label")
}

func (p *plastic) parseBranchSelector(selector string) string {
	return p.parseStringPattern(selector, `branch "(?P<branch>[\/a-zA-Z0-9\-\_]+?)"`, "branch")
}

func (p *plastic) getCmCommandOutput(args ...string) string {
	val, _ := p.env.RunCommand("cm", args...)
	return val
}
