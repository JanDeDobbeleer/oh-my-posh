package main

import (
	"bytes"
	"fmt"
)

const (
	// LocalWorkingIcon the icon to use as the local working area changes indicator
	LocalWorkingIcon Property = "local_working_icon"
	// LocalStagingIcon the icon to use as the local staging area changes indicator
	LocalStagingIcon Property = "local_staged_icon"
	// DisplayStatusDetail shows the detailed status of the repository
	DisplayStatusDetail Property = "display_status_detail"
	// WorkingColor if set, the color to use on the working area
	WorkingColor Property = "working_color"
	// StagingColor if set, the color to use on the staging area
	StagingColor Property = "staging_color"
	// StatusColorsEnabled enables status colors
	StatusColorsEnabled Property = "status_colors_enabled"
	// LocalChangesColor if set, the color to use when there are local changes
	LocalChangesColor Property = "local_changes_color"
	// AheadAndBehindColor if set, the color to use when the branch is ahead and behind the remote
	AheadAndBehindColor Property = "ahead_and_behind_color"
	// BehindColor if set, the color to use when the branch is ahead and behind the remote
	BehindColor Property = "behind_color"
	// AheadColor if set, the color to use when the branch is ahead and behind the remote
	AheadColor Property = "ahead_color"
	// BranchMaxLength truncates the length of the branch name
	BranchMaxLength Property = "branch_max_length"
	// WorktreeCountIcon shows before the worktree context
	WorktreeCountIcon Property = "worktree_count_icon"
)

func (g *git) renderDeprecatedString(displayStatus bool) string {
	if !displayStatus {
		return g.getPrettyHEADName()
	}
	buffer := new(bytes.Buffer)
	// remote (if available)
	if len(g.repo.UpstreamIcon) != 0 {
		fmt.Fprintf(buffer, "%s", g.repo.UpstreamIcon)
	}
	// branchName
	fmt.Fprintf(buffer, "%s", g.repo.HEAD)
	if g.props.getBool(DisplayBranchStatus, true) {
		buffer.WriteString(g.getBranchStatus())
	}
	if g.repo.Staging.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.Staging, StagingColor, LocalStagingIcon, " \uF046"))
	}
	if g.repo.Staging.Changed && g.repo.Working.Changed {
		fmt.Fprint(buffer, g.props.getString(StatusSeparatorIcon, " |"))
	}
	if g.repo.Working.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.Working, WorkingColor, LocalWorkingIcon, " \uF044"))
	}
	if g.repo.StashCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(StashCountIcon, "\uF692 "), g.repo.StashCount)
	}
	if g.repo.WorktreeCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(WorktreeCountIcon, "\uf1bb "), g.repo.WorktreeCount)
	}
	return buffer.String()
}
