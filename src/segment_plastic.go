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

	plasticWorkspaceFolder string // .plastic working folder

	cmCommand string
}

const (
	// FullBranchPath displays the full path of a branch
	FullBranchPath Property = "full_branch_path"
)

func (g *plastic) enabled() bool {
	if !g.env.hasCommand(g.getCmCommand()) {
		return false
	}
	wkdir, err := g.env.hasParentFilePath(".plastic")
	if err != nil {
		return false
	}
	if g.shouldIgnoreRootRepository(wkdir.parentFolder) {
		return false
	}

	if wkdir.isDir {
		g.plasticWorkspaceFolder = wkdir.parentFolder
		return true
	}

	return false
}

func (g *plastic) shouldIgnoreRootRepository(rootDir string) bool {
	value, ok := g.props[ExcludeFolders]
	if !ok {
		return false
	}
	excludedFolders := parseStringArray(value)
	return dirMatchesOneOf(g.env, rootDir, excludedFolders)
}

func (g *plastic) string() string {
	statusColorsEnabled := g.props.getBool(StatusColorsEnabled, false)
	displayStatus := g.props.getOneOfBool(FetchStatus, DisplayStatus, false)

	g.Selector = g.getSelector()
	if displayStatus || statusColorsEnabled {
		g.getPlasticStatus()
	}

	// use template if available
	segmentTemplate := g.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) > 0 {
		return g.templateString(segmentTemplate)
	}

	// default: only selector is returned
	return g.Selector
}

func (g *plastic) templateString(segmentTemplate string) string {
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  g,
		Env:      g.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (g *plastic) Status() string {
	var status string
	stringIfValue := func(value int, prefix string) string {
		if value > 0 {
			return fmt.Sprintf(" %s%d", prefix, value)
		}
		return ""
	}
	status += stringIfValue(g.Added, "+")
	status += stringIfValue(g.Modified, "~")
	status += stringIfValue(g.Deleted, "-")
	status += stringIfValue(g.Moved, ">")
	status += stringIfValue(g.Unmerged, "x")
	return strings.TrimSpace(status)
}

func (g *plastic) getPlasticStatus() {
	output := g.getCmCommandOutput("status", "--all", "--machinereadable")
	splittedOutput := strings.Split(output, "\n")
	// compare to head
	currentChangeset := g.parseStatusChangeset(splittedOutput[0])
	headChangeset := g.getHeadChangeset()
	g.Behind = headChangeset > currentChangeset

	// parse file state
	g.parseFilesStatus(splittedOutput)
}

func (g *plastic) parseFilesStatus(output []string) {
	if len(output) <= 1 {
		return
	}
	for _, line := range output[1:] {
		if len(line) < 3 {
			continue
		}
		if matchString(`(?i)merge\s+from\s+[0-9]+\s*$`, line) {
			g.Unmerged++
			continue
		}
		code := line[0:2]
		switch code {
		case "LD":
			g.Deleted++
		case "AD", "PR":
			g.Added++
		case "LM":
			g.Moved++
		case "CH":
			g.Modified++
		}
	}
	g.Changed = g.Added > 0 || g.Deleted > 0 || g.Modified > 0 || g.Moved > 0 || g.Unmerged > 0
}

func (g *plastic) parseStatusChangeset(status string) int {
	var csRegex = `STATUS\s+(?P<cs>[0-9]+?)\s`
	cs, _ := strconv.Atoi(findNamedRegexMatch(csRegex, status)["cs"])
	return cs
}

func (g *plastic) getHeadChangeset() int {
	output := g.getCmCommandOutput("status", "--head", "--machinereadable")
	var csRegex = `\bcs:(?P<cs>[0-9]+?)\s`
	cs, _ := strconv.Atoi(findNamedRegexMatch(csRegex, output)["cs"])
	return cs
}

func (g *plastic) getPlasticFileContents(folder, file string) string {
	return strings.Trim(g.env.getFileContent(folder+"/"+file), " \r\n")
}

func (g *plastic) getSelector() string {
	var ref string
	selector := g.getPlasticFileContents(g.plasticWorkspaceFolder+"/.plastic", "plastic.selector")
	// changeset
	ref = g.parseChangesetSelector(selector)
	if ref != "" {
		return fmt.Sprintf("%s%s", g.props.getString(CommitIcon, "\uF417"), ref)
	}
	// fallback to label
	ref = g.parseLabelSelector(selector)
	if ref != "" {
		return fmt.Sprintf("%s%s", g.props.getString(TagIcon, "\uF412"), ref)
	}
	// fallback to branch/smartbranch
	ref = g.parseBranchSelector(selector)
	if ref != "" {
		ref = g.truncateBranch(ref)
	}
	return fmt.Sprintf("%s%s", g.props.getString(BranchIcon, "\uE0A0"), ref)
}

func (g *plastic) parseChangesetSelector(selector string) string {
	var csRegex = `\bchangeset "(?P<cs>[0-9]+?)"`
	return findNamedRegexMatch(csRegex, selector)["cs"]
}

func (g *plastic) parseLabelSelector(selector string) string {
	var labelRegex = `label "(?P<label>[a-zA-Z0-9\-\_]+?)"`
	return findNamedRegexMatch(labelRegex, selector)["label"]
}

func (g *plastic) parseBranchSelector(selector string) string {
	var branchRegex = `branch "(?P<branch>[\/a-zA-Z0-9\-\_]+?)"`
	return findNamedRegexMatch(branchRegex, selector)["branch"]
}

func (g *plastic) init(props properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *plastic) truncateBranch(branch string) string {
	fullBranchPath := g.props.getBool(FullBranchPath, false)
	maxLength := g.props.getInt(BranchMaxLength, 0)
	if !fullBranchPath && len(branch) > 0 {
		index := strings.LastIndex(branch, "/")
		branch = branch[index+1:]
	}
	if maxLength == 0 || len(branch) <= maxLength {
		return branch
	}
	symbol := g.props.getString(TruncateSymbol, "")
	return branch[0:maxLength] + symbol
}

func (g *plastic) getCmCommandOutput(args ...string) string {
	val, _ := g.env.runCommand(g.getCmCommand(), args...)
	return val
}

func (g *plastic) getCmCommand() string {
	if len(g.cmCommand) > 0 {
		return g.cmCommand
	}
	inWSL2SharedDrive := func(env environmentInfo) bool {
		if !env.isWsl() {
			return false
		}
		if !strings.HasPrefix(env.getcwd(), "/mnt/") {
			return false
		}
		uname, _ := g.env.runCommand("uname", "-r")
		return strings.Contains(uname, "WSL2")
	}
	g.cmCommand = "cm"
	if g.env.getRuntimeGOOS() == windowsPlatform || inWSL2SharedDrive(g.env) {
		g.cmCommand = "cm.exe"
	}
	return g.cmCommand
}
