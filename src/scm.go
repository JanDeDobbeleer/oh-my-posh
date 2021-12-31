package main

import (
	"fmt"
	"strings"
)

// ScmStatus represents part of the status of a repository
type ScmStatus struct {
	Unmerged int
	Deleted  int
	Added    int
	Modified int
	Moved    int
}

func (s *ScmStatus) Changed() bool {
	return s.Added > 0 || s.Deleted > 0 || s.Modified > 0 || s.Unmerged > 0 || s.Moved > 0
}

func (s *ScmStatus) String() string {
	var status string
	stringIfValue := func(value int, prefix string) string {
		if value > 0 {
			return fmt.Sprintf(" %s%d", prefix, value)
		}
		return ""
	}
	status += stringIfValue(s.Added, "+")
	status += stringIfValue(s.Modified, "~")
	status += stringIfValue(s.Deleted, "-")
	status += stringIfValue(s.Moved, ">")
	status += stringIfValue(s.Unmerged, "x")
	return strings.TrimSpace(status)
}

type scm struct {
	props properties
	env   environmentInfo
}

const (
	// BranchMaxLength truncates the length of the branch name
	BranchMaxLength Property = "branch_max_length"
	// TruncateSymbol appends the set symbol to a truncated branch name
	TruncateSymbol Property = "truncate_symbol"
	// FullBranchPath displays the full path of a branch
	FullBranchPath Property = "full_branch_path"
)

func (s *scm) init(props properties, env environmentInfo) {
	s.props = props
	s.env = env
}

func (s *scm) truncateBranch(branch string) string {
	fullBranchPath := s.props.getBool(FullBranchPath, true)
	maxLength := s.props.getInt(BranchMaxLength, 0)
	if !fullBranchPath && strings.Contains(branch, "/") {
		index := strings.LastIndex(branch, "/")
		branch = branch[index+1:]
	}
	if maxLength == 0 || len(branch) <= maxLength {
		return branch
	}
	symbol := s.props.getString(TruncateSymbol, "")
	return branch[0:maxLength] + symbol
}

func (s *scm) shouldIgnoreRootRepository(rootDir string) bool {
	value, ok := s.props[ExcludeFolders]
	if !ok {
		return false
	}
	excludedFolders := parseStringArray(value)
	return dirMatchesOneOf(s.env, rootDir, excludedFolders)
}

func (s *scm) getFileContents(folder, file string) string {
	return strings.Trim(s.env.getFileContent(folder+"/"+file), " \r\n")
}
