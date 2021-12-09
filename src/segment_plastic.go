package main

import (
	"fmt"
	"strconv"
	"strings"
)

type plastic struct {
	props properties
	env   environmentInfo

	Moved    int
	Deleted  int
	Added    int
	Modified int
	Unmerged int
	Changed  bool

	Behind   bool
	Selector string

	plasticWorkspaceFolder string // root folder of workspace
}

const (
	// FullBranchPath displays the full path of a branch
	FullBranchPath Property = "full_branch_path"
)

func (p *plastic) enabled() bool {
	if !p.env.hasCommand("cm") {
		return false
	}
	wkdir, err := p.env.hasParentFilePath(".plastic")
	if err != nil {
		return false
	}
	if p.shouldIgnoreRootRepository(wkdir.parentFolder) {
		return false
	}

	if wkdir.isDir {
		p.plasticWorkspaceFolder = wkdir.parentFolder
		return true
	}

	return false
}

func (p *plastic) shouldIgnoreRootRepository(rootDir string) bool {
	value, ok := p.props[ExcludeFolders]
	if !ok {
		return false
	}
	excludedFolders := parseStringArray(value)
	return dirMatchesOneOf(p.env, rootDir, excludedFolders)
}

func (p *plastic) string() string {
	displayStatus := p.props.getOneOfBool(FetchStatus, DisplayStatus, false)

	p.Selector = p.getSelector()
	if displayStatus {
		p.getPlasticStatus()
	}

	// use template if available
	segmentTemplate := p.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) > 0 {
		return p.templateString(segmentTemplate)
	}

	// default: only selector is returned
	return p.Selector
}

func (p *plastic) templateString(segmentTemplate string) string {
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

func (p *plastic) Status() string {
	var status string
	stringIfValue := func(value int, prefix string) string {
		if value > 0 {
			return fmt.Sprintf(" %s%d", prefix, value)
		}
		return ""
	}
	status += stringIfValue(p.Added, "+")
	status += stringIfValue(p.Modified, "~")
	status += stringIfValue(p.Deleted, "-")
	status += stringIfValue(p.Moved, ">")
	status += stringIfValue(p.Unmerged, "x")
	return strings.TrimSpace(status)
}

func (p *plastic) getPlasticStatus() {
	output := p.getCmCommandOutput("status", "--all", "--machinereadable")
	splittedOutput := strings.Split(output, "\n")
	// compare to head
	currentChangeset := p.parseStatusChangeset(splittedOutput[0])
	headChangeset := p.getHeadChangeset()
	p.Behind = headChangeset > currentChangeset

	// parse file state
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
		if matchString(`(?i)merge\s+from\s+[0-9]+\s*$`, line) {
			p.Unmerged++
			continue
		}
		code := line[0:2]
		switch code {
		case "LD":
			p.Deleted++
		case "AD", "PR":
			p.Added++
		case "LM":
			p.Moved++
		case "CH":
			p.Modified++
		}
	}
	p.Changed = p.Added > 0 || p.Deleted > 0 || p.Modified > 0 || p.Moved > 0 || p.Unmerged > 0
}

func (p *plastic) parseStringPattern(output, pattern, name string) string {
	match := findNamedRegexMatch(pattern, output)
	if sValue, ok := match[name]; ok {
		return sValue
	}
	return ""
}

func (p *plastic) parseIntPattern(output, pattern, name string, defValue int) int {
	sValue := p.parseStringPattern(output, pattern, name)
	if sValue != "" {
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

func (p *plastic) getPlasticFileContents(file string) string {
	return strings.Trim(p.env.getFileContent(p.plasticWorkspaceFolder+"/.plastic/"+file), " \r\n")
}

func (p *plastic) getSelector() string {
	var ref string
	selector := p.getPlasticFileContents("plastic.selector")
	// changeset
	ref = p.parseChangesetSelector(selector)
	if ref != "" {
		return fmt.Sprintf("%s%s", p.props.getString(CommitIcon, "\uF417"), ref)
	}
	// fallback to label
	ref = p.parseLabelSelector(selector)
	if ref != "" {
		return fmt.Sprintf("%s%s", p.props.getString(TagIcon, "\uF412"), ref)
	}
	// fallback to branch/smartbranch
	ref = p.parseBranchSelector(selector)
	if ref != "" {
		ref = p.truncateBranch(ref)
	}
	return fmt.Sprintf("%s%s", p.props.getString(BranchIcon, "\uE0A0"), ref)
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

func (p *plastic) init(props properties, env environmentInfo) {
	p.props = props
	p.env = env
}

func (p *plastic) truncateBranch(branch string) string {
	fullBranchPath := p.props.getBool(FullBranchPath, false)
	maxLength := p.props.getInt(BranchMaxLength, 0)
	if !fullBranchPath && len(branch) > 0 {
		index := strings.LastIndex(branch, "/")
		branch = branch[index+1:]
	}
	if maxLength == 0 || len(branch) <= maxLength {
		return branch
	}
	symbol := p.props.getString(TruncateSymbol, "")
	return branch[0:maxLength] + symbol
}

func (p *plastic) getCmCommandOutput(args ...string) string {
	val, _ := p.env.runCommand("cm", args...)
	return val
}
