package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/distatus/battery"
)

// Segment

const (
	BackgroundOverride Property = "background"
	ForegroundOverride Property = "foreground"
)

// Properties

func (p properties) getOneOfBool(property, legacyProperty Property) bool {
	_, found := p[legacyProperty]
	if found {
		return p.getBool(legacyProperty, false)
	}
	return p.getBool(property, false)
}

func (p properties) hasOneOf(properties ...Property) bool {
	for _, property := range properties {
		if _, found := p[property]; found {
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
		g.props[BackgroundOverride] = g.getStatusColor(g.props.getColor(BackgroundOverride, ""))
	} else {
		g.props[ForegroundOverride] = g.getStatusColor(g.props.getColor(ForegroundOverride, ""))
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
	foregroundColor := g.props.getColor(color, g.props.getColor(ForegroundOverride, ""))
	detail := ""
	if g.props.getBool(DisplayStatusDetail, true) {
		detail = status.String()
	}
	statusStr := g.colorStatusString(prefix, detail, foregroundColor)
	return strings.TrimSpace(statusStr)
}

func (g *git) colorStatusString(prefix, status, color string) string {
	if len(color) == 0 {
		return fmt.Sprintf("%s %s", prefix, status)
	}
	if len(status) == 0 {
		return fmt.Sprintf("<%s>%s</>", color, prefix)
	}
	if strings.Contains(prefix, "</>") {
		return fmt.Sprintf("%s <%s>%s</>", prefix, color, status)
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
		e.props[ForegroundOverride] = e.props.getColor(ErrorColor, e.props.getColor(ForegroundOverride, ""))
	}
	if e.Code != 0 && colorBackground {
		e.props[BackgroundOverride] = e.props.getColor(ErrorColor, e.props.getColor(BackgroundOverride, ""))
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
		b.props[BackgroundOverride] = b.props.getColor(colorProperty, b.props.getColor(BackgroundOverride, ""))
	} else {
		b.props[ForegroundOverride] = b.props.getColor(colorProperty, b.props.getColor(ForegroundOverride, ""))
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

// Session

const (
	// UserInfoSeparator is put between the user and computer name
	UserInfoSeparator Property = "user_info_separator"
	// UserColor if set, is used to color the user name
	UserColor Property = "user_color"
	// HostColor if set, is used to color the computer name
	HostColor Property = "host_color"
	// DisplayHost hides or show the computer name
	DisplayHost Property = "display_host"
	// DisplayUser hides or shows the user name
	DisplayUser Property = "display_user"
	// DefaultUserName holds the default user of the platform
	DefaultUserName Property = "default_user_name"
	// SSHIcon is the icon used for SSH sessions
	SSHIcon Property = "ssh_icon"

	defaultUserEnvVar = "POSH_SESSION_DEFAULT_USER"
)

func (s *session) getDefaultUser() string {
	user := s.env.getenv(defaultUserEnvVar)
	if len(user) == 0 {
		user = s.props.getString(DefaultUserName, "")
	}
	return user
}

func (s *session) legacyEnabled() bool {
	if s.props.getBool(DisplayUser, true) {
		s.UserName = s.getUserName()
	}
	if s.props.getBool(DisplayHost, true) {
		s.ComputerName = s.getComputerName()
	}
	s.DefaultUserName = s.getDefaultUser()
	showDefaultUser := s.props.getBool(DisplayDefault, true)
	if !showDefaultUser && s.DefaultUserName == s.UserName {
		return false
	}
	return true
}

func (s *session) getFormattedText() string {
	separator := ""
	if s.props.getBool(DisplayHost, true) && s.props.getBool(DisplayUser, true) {
		separator = s.props.getString(UserInfoSeparator, "@")
	}
	var sshIcon string
	if s.SSHSession {
		sshIcon = s.props.getString(SSHIcon, "\uF817 ")
	}
	defaulColor := s.props.getColor(ForegroundOverride, "")
	userColor := s.props.getColor(UserColor, defaulColor)
	hostColor := s.props.getColor(HostColor, defaulColor)
	if len(userColor) > 0 && len(hostColor) > 0 {
		return fmt.Sprintf("%s<%s>%s</>%s<%s>%s</>", sshIcon, userColor, s.UserName, separator, hostColor, s.ComputerName)
	}
	if len(userColor) > 0 {
		return fmt.Sprintf("%s<%s>%s</>%s%s", sshIcon, userColor, s.UserName, separator, s.ComputerName)
	}
	if len(hostColor) > 0 {
		return fmt.Sprintf("%s%s%s<%s>%s</>", sshIcon, s.UserName, separator, hostColor, s.ComputerName)
	}
	return fmt.Sprintf("%s%s%s%s", sshIcon, s.UserName, separator, s.ComputerName)
}
