package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/distatus/battery"
)

// Properties

func (p *properties) getOneOfBool(property, legacyProperty Property) bool {
	_, found := p.values[legacyProperty]
	if found {
		return p.getBool(legacyProperty, false)
	}
	return p.getBool(property, false)
}

func (p *properties) hasOneOf(properties ...Property) bool {
	for _, property := range properties {
		if _, found := p.values[property]; found {
			return true
		}
	}
	return false
}

// GIT Segement

const (
	// DisplayStatus shows the status of the repository
	DisplayStatus Property = "display_status"
	// DisplayStashCount show stash count or not
	DisplayStashCount Property = "display_stash_count"
	// DisplayWorktreeCount show worktree count or not
	DisplayWorktreeCount Property = "display_worktree_count"
	// DisplayUpstreamIcon show or hide the upstream icon
	DisplayUpstreamIcon Property = "display_upstream_icon"
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
	// WorktreeCountIcon shows before the worktree context
	WorktreeCountIcon Property = "worktree_count_icon"
	// StashCountIcon shows before the stash context
	StashCountIcon Property = "stash_count_icon"
	// StatusSeparatorIcon shows between staging and working area
	StatusSeparatorIcon Property = "status_separator_icon"
)

func (g *git) deprecatedString(statusColorsEnabled bool) string {
	if statusColorsEnabled {
		g.SetStatusColor()
	}
	buffer := new(bytes.Buffer)
	// remote (if available)
	if len(g.UpstreamIcon) != 0 {
		fmt.Fprintf(buffer, "%s", g.UpstreamIcon)
	}
	// branchName
	fmt.Fprintf(buffer, "%s", g.HEAD)
	if len(g.BranchStatus) > 0 {
		buffer.WriteString(g.BranchStatus)
	}
	if g.Staging.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.Staging, StagingColor, LocalStagingIcon, " \uF046"))
	}
	if g.Staging.Changed && g.Working.Changed {
		fmt.Fprint(buffer, g.props.getString(StatusSeparatorIcon, " |"))
	}
	if g.Working.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.Working, WorkingColor, LocalWorkingIcon, " \uF044"))
	}
	if g.StashCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(StashCountIcon, "\uF692 "), g.StashCount)
	}
	if g.WorktreeCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(WorktreeCountIcon, "\uf1bb "), g.WorktreeCount)
	}
	return buffer.String()
}

func (g *git) SetStatusColor() {
	if g.props.getBool(ColorBackground, true) {
		g.props.background = g.getStatusColor(g.props.background)
	} else {
		g.props.foreground = g.getStatusColor(g.props.foreground)
	}
}

func (g *git) getStatusColor(defaultValue string) string {
	if g.Staging.Changed || g.Working.Changed {
		return g.props.getColor(LocalChangesColor, defaultValue)
	} else if g.Ahead > 0 && g.Behind > 0 {
		return g.props.getColor(AheadAndBehindColor, defaultValue)
	} else if g.Ahead > 0 {
		return g.props.getColor(AheadColor, defaultValue)
	} else if g.Behind > 0 {
		return g.props.getColor(BehindColor, defaultValue)
	}
	return defaultValue
}

func (g *git) getStatusDetailString(status *GitStatus, color, icon Property, defaultIcon string) string {
	prefix := g.props.getString(icon, defaultIcon)
	foregroundColor := g.props.getColor(color, g.props.foreground)
	if !g.props.getBool(DisplayStatusDetail, true) {
		return g.colorStatusString(prefix, "", foregroundColor)
	}
	return g.colorStatusString(prefix, status.String(), foregroundColor)
}

func (g *git) colorStatusString(prefix, status, color string) string {
	if color == g.props.foreground && len(status) == 0 {
		return prefix
	}
	if color == g.props.foreground {
		return fmt.Sprintf("%s %s", prefix, status)
	}
	if strings.Contains(prefix, "</>") {
		return fmt.Sprintf("%s <%s>%s</>", prefix, color, status)
	}
	if len(status) == 0 {
		return fmt.Sprintf("<%s>%s</>", color, prefix)
	}
	return fmt.Sprintf("<%s>%s %s</>", color, prefix, status)
}

// EXIT Segment

const (
	// DisplayExitCode shows or hides the error code
	DisplayExitCode Property = "display_exit_code"
	// ErrorColor specify a different foreground color for the error text when using always_show = true
	ErrorColor Property = "error_color"
	// AlwaysNumeric shows error codes as numbers
	AlwaysNumeric Property = "always_numeric"
	// SuccessIcon displays when there's no error and AlwaysEnabled = true
	SuccessIcon Property = "success_icon"
	// ErrorIcon displays when there's an error
	ErrorIcon Property = "error_icon"
)

func (e *exit) deprecatedString() string {
	colorBackground := e.props.getBool(ColorBackground, false)
	if e.Code != 0 && !colorBackground {
		e.props.foreground = e.props.getColor(ErrorColor, e.props.foreground)
	}
	if e.Code != 0 && colorBackground {
		e.props.background = e.props.getColor(ErrorColor, e.props.background)
	}
	if e.Code == 0 {
		return e.props.getString(SuccessIcon, "")
	}
	errorIcon := e.props.getString(ErrorIcon, "")
	if !e.props.getBool(DisplayExitCode, true) {
		return errorIcon
	}
	if e.props.getBool(AlwaysNumeric, false) {
		return fmt.Sprintf("%s%d", errorIcon, e.Code)
	}
	return fmt.Sprintf("%s%s", errorIcon, e.Text)
}

// Battery segment

const (
	// ChargedColor to display when fully charged
	ChargedColor Property = "charged_color"
	// ChargingColor to display when charging
	ChargingColor Property = "charging_color"
	// DischargingColor to display when discharging
	DischargingColor Property = "discharging_color"
	// DisplayCharging Hide the battery icon while it's charging
	DisplayCharging Property = "display_charging"
	// DisplayCharged Hide the battery icon when it's charged
	DisplayCharged Property = "display_charged"
)

func (b *batt) colorSegment() {
	if !b.props.hasOneOf(ChargedColor, ChargingColor, DischargingColor) {
		return
	}
	var colorProperty Property
	switch b.Battery.State {
	case battery.Discharging, battery.NotCharging:
		colorProperty = DischargingColor
	case battery.Charging:
		colorProperty = ChargingColor
	case battery.Full:
		colorProperty = ChargedColor
	case battery.Empty, battery.Unknown:
		return
	}
	colorBackground := b.props.getBool(ColorBackground, false)
	if colorBackground {
		b.props.background = b.props.getColor(colorProperty, b.props.background)
	} else {
		b.props.foreground = b.props.getColor(colorProperty, b.props.foreground)
	}
}

func (b *batt) shouldDisplay() bool {
	if !b.props.hasOneOf(DisplayCharged, DisplayCharging) {
		return true
	}
	displayCharged := b.props.getBool(DisplayCharged, true)
	if !displayCharged && (b.Battery.State == battery.Full) {
		return false
	}
	displayCharging := b.props.getBool(DisplayCharging, true)
	if !displayCharging && (b.Battery.State == battery.Charging) {
		return false
	}
	return true
}
