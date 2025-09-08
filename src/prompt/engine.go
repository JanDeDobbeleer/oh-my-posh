package prompt

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

var cycle *color.Cycle = &color.Cycle{}

type Engine struct {
	Env                   runtime.Environment
	Config                *config.Config
	activeSegment         *config.Segment
	previousActiveSegment *config.Segment
	rprompt               string
	Overflow              config.Overflow
	prompt                strings.Builder
	rpromptLength         int
	Padding               int
	currentLineLength     int
	Plain                 bool
	forceRender           bool
}

const (
	PRIMARY   = "primary"
	TRANSIENT = "transient"
	DEBUG     = "debug"
	SECONDARY = "secondary"
	RIGHT     = "right"
	TOOLTIP   = "tooltip"
	VALID     = "valid"
	ERROR     = "error"
	PREVIEW   = "preview"
)

func (e *Engine) write(txt string) {
	// Grow capacity proactively if needed
	if e.prompt.Cap() < e.prompt.Len()+len(txt) {
		e.prompt.Grow(len(txt) * 2) // Grow by double the needed size to reduce future allocations
	}
	e.prompt.WriteString(txt)
}

func (e *Engine) string() string {
	txt := e.prompt.String()
	e.prompt.Reset()
	return txt
}

func (e *Engine) canWriteRightBlock(length int, rprompt bool) (int, bool) {
	if rprompt && (e.rprompt == "") {
		return 0, false
	}

	consoleWidth, err := e.Env.TerminalWidth()
	if err != nil || consoleWidth == 0 {
		return 0, false
	}

	availableSpace := consoleWidth - e.currentLineLength

	// spanning multiple lines
	if availableSpace < 0 {
		overflow := e.currentLineLength % consoleWidth
		availableSpace = consoleWidth - overflow
	}

	availableSpace -= length

	promptBreathingRoom := 5
	if rprompt {
		promptBreathingRoom = 30
	}

	canWrite := availableSpace >= promptBreathingRoom

	// reset the available space when we can't write so we can fill the line
	if !canWrite {
		availableSpace = consoleWidth - length
	}

	return availableSpace, canWrite
}

func (e *Engine) pwd() {
	// only print when relevant
	if e.Config.PWD == "" {
		return
	}

	// only print when supported
	sh := e.Env.Shell()
	if sh == shell.ELVISH || sh == shell.XONSH {
		return
	}

	pwd := e.Env.Pwd()
	if e.Env.IsCygwin() {
		pwd = strings.ReplaceAll(pwd, `\`, `/`)
	}

	// Allow template logic to define when to enable the PWD (when supported)
	pwdType, err := template.Render(e.Config.PWD, nil)
	if err != nil || pwdType == "" {
		return
	}

	user := e.Env.User()
	host, _ := e.Env.Host()
	e.write(terminal.Pwd(pwdType, user, host, pwd))
}

func (e *Engine) getNewline() string {
	newline := "\n"

	if e.Plain || e.Env.Flags().Debug {
		return newline
	}

	// Warp terminal will remove a newline character ('\n') from the prompt, so we hack it in.
	if e.isWarp() {
		return terminal.LineBreak()
	}

	return newline
}

func (e *Engine) writeNewline() {
	defer func() {
		e.currentLineLength = 0
	}()

	e.write(e.getNewline())
}

func (e *Engine) isWarp() bool {
	return terminal.Program == terminal.Warp
}

func (e *Engine) isIterm() bool {
	return terminal.Program == terminal.ITerm
}

func (e *Engine) shouldFill(filler string, padLength int) (string, bool) {
	if filler == "" {
		return "", false
	}

	var err error
	if filler, err = template.Render(filler, e); err != nil {
		return "", false
	}

	// allow for easy color overrides and templates
	terminal.SetColors("default", "default")
	terminal.Write("", "", filler)
	filler, lenFiller := terminal.String()
	if lenFiller == 0 {
		return "", false
	}

	repeat := padLength / lenFiller
	unfilled := padLength % lenFiller
	txt := strings.Repeat(filler, repeat) + strings.Repeat(" ", unfilled)
	return txt, true
}

func (e *Engine) getTitleTemplateText() string {
	if txt, err := template.Render(e.Config.ConsoleTitleTemplate, nil); err == nil {
		return txt
	}

	return ""
}

func (e *Engine) renderBlock(block *config.Block, cancelNewline bool) bool {
	txt, length := e.writeBlockSegments(block)

	// do not print anything when we don't have any text unless forced
	if !block.Force && length == 0 {
		return false
	}

	defer e.applyPowerShellBleedPatch()

	// do not print a newline to avoid a leading space
	// when we're printing the first primary prompt in
	// the shell
	if block.Newline && !cancelNewline {
		e.writeNewline()
	}

	switch block.Type {
	case config.Prompt:
		if block.Alignment == config.Left {
			e.currentLineLength += length
			e.write(txt)
			return true
		}

		if block.Alignment != config.Right {
			return false
		}

		space, OK := e.canWriteRightBlock(length, false)

		// we can't print the right block as there's not enough room available
		if !OK {
			e.Overflow = block.Overflow
			switch block.Overflow {
			case config.Break:
				e.writeNewline()
			case config.Hide:
				// make sure to fill if needed
				if padText, OK := e.shouldFill(block.Filler, space+length-e.currentLineLength); OK {
					e.write(padText)
				}

				e.currentLineLength = 0
				return true
			}
		}

		defer func() {
			e.currentLineLength = 0
			e.Overflow = ""
		}()

		// validate if we have a filler and fill if needed
		if padText, OK := e.shouldFill(block.Filler, space); OK {
			e.write(padText)
			e.write(txt)
			return true
		}

		if space > 0 {
			e.write(strings.Repeat(" ", space))
		}

		e.write(txt)
	case config.RPrompt:
		e.rprompt = txt
		e.rpromptLength = length
	}

	return true
}

func (e *Engine) applyPowerShellBleedPatch() {
	// when in PowerShell, we need to clear the line after the prompt
	// to avoid the background being printed on the next line
	// when at the end of the buffer.
	// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
	if e.Env.Shell() != shell.PWSH && e.Env.Shell() != shell.PWSH5 {
		return
	}

	// only do this when enabled
	if !e.Config.PatchPwshBleed {
		return
	}

	e.write(terminal.ClearAfter())
}

func (e *Engine) setActiveSegment(segment *config.Segment) {
	e.activeSegment = segment
	terminal.Interactive = segment.Interactive
	terminal.SetColors(segment.ResolveBackground(), segment.ResolveForeground())
}

func (e *Engine) renderActiveSegment() {
	e.writeSeparator(false)

	switch e.activeSegment.ResolveStyle() {
	case config.Plain, config.Powerline:
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Diamond:
		background := color.Transparent

		if e.previousActiveSegment != nil && e.previousActiveSegment.HasEmptyDiamondAtEnd() {
			background = e.previousActiveSegment.ResolveBackground()
		}

		terminal.Write(background, color.Background, e.activeSegment.LeadingDiamond)
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Accordion:
		if e.activeSegment.Enabled {
			terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
		}
	}

	e.previousActiveSegment = e.activeSegment

	terminal.SetParentColors(e.previousActiveSegment.ResolveBackground(), e.previousActiveSegment.ResolveForeground())
}

func (e *Engine) writeSeparator(final bool) {
	if e.activeSegment == nil {
		return
	}

	isCurrentDiamond := e.activeSegment.ResolveStyle() == config.Diamond
	if final && isCurrentDiamond {
		terminal.Write(color.Transparent, color.Background, e.activeSegment.TrailingDiamond)
		return
	}

	isPreviousDiamond := e.previousActiveSegment != nil && e.previousActiveSegment.ResolveStyle() == config.Diamond
	if isPreviousDiamond {
		e.adjustTrailingDiamondColorOverrides()
	}

	if isPreviousDiamond && isCurrentDiamond && e.activeSegment.LeadingDiamond == "" {
		terminal.Write(color.Background, color.ParentBackground, e.previousActiveSegment.TrailingDiamond)
		return
	}

	if isPreviousDiamond && len(e.previousActiveSegment.TrailingDiamond) > 0 {
		terminal.Write(color.Transparent, color.ParentBackground, e.previousActiveSegment.TrailingDiamond)
	}

	isPowerline := e.activeSegment.IsPowerline()

	shouldOverridePowerlineLeadingSymbol := func() bool {
		if !isPowerline {
			return false
		}

		if isPowerline && e.activeSegment.LeadingPowerlineSymbol == "" {
			return false
		}

		if e.previousActiveSegment != nil && e.previousActiveSegment.IsPowerline() {
			return false
		}

		return true
	}

	if shouldOverridePowerlineLeadingSymbol() {
		terminal.Write(color.Transparent, color.Background, e.activeSegment.LeadingPowerlineSymbol)
		return
	}

	resolvePowerlineSymbol := func() string {
		if isPowerline {
			return e.activeSegment.PowerlineSymbol
		}

		if e.previousActiveSegment != nil && e.previousActiveSegment.IsPowerline() {
			return e.previousActiveSegment.PowerlineSymbol
		}

		return ""
	}

	symbol := resolvePowerlineSymbol()
	if symbol == "" {
		return
	}

	bgColor := color.Background
	if final || !isPowerline {
		bgColor = color.Transparent
	}

	if e.activeSegment.ResolveStyle() == config.Diamond && e.activeSegment.LeadingDiamond == "" {
		bgColor = color.Background
	}

	if e.activeSegment.InvertPowerline || (e.previousActiveSegment != nil && e.previousActiveSegment.InvertPowerline) {
		terminal.Write(e.getPowerlineColor(), bgColor, symbol)
		return
	}

	terminal.Write(bgColor, e.getPowerlineColor(), symbol)
}

func (e *Engine) getPowerlineColor() color.Ansi {
	if e.previousActiveSegment == nil {
		return color.Transparent
	}

	if e.previousActiveSegment.ResolveStyle() == config.Diamond && e.previousActiveSegment.TrailingDiamond == "" {
		return e.previousActiveSegment.ResolveBackground()
	}

	if e.activeSegment.ResolveStyle() == config.Diamond && e.activeSegment.LeadingDiamond == "" {
		return e.previousActiveSegment.ResolveBackground()
	}

	if !e.previousActiveSegment.IsPowerline() {
		return color.Transparent
	}

	return e.previousActiveSegment.ResolveBackground()
}

func (e *Engine) adjustTrailingDiamondColorOverrides() {
	// as we now already adjusted the activeSegment, we need to change the value
	// of background and foreground to parentBackground and parentForeground
	// this will still break when using parentBackground and parentForeground as keywords
	// in a trailing diamond, but let's fix that when it happens as it requires either a rewrite
	// of the logic for diamonds or storing grandparents as well like one happy family.
	if e.previousActiveSegment == nil || e.previousActiveSegment.TrailingDiamond == "" {
		return
	}

	trailingDiamond := e.previousActiveSegment.TrailingDiamond
	// Optimize: check both conditions in a single pass
	hasBg := strings.Contains(trailingDiamond, string(color.Background))
	hasFg := strings.Contains(trailingDiamond, string(color.Foreground))

	if !hasBg && !hasFg {
		return
	}

	match := regex.FindNamedRegexMatch(terminal.AnchorRegex, trailingDiamond)
	if len(match) == 0 {
		return
	}

	adjustOverride := func(anchor string, override color.Ansi) {
		newOverride := override
		switch override { //nolint:exhaustive
		case color.Foreground:
			newOverride = color.ParentForeground
		case color.Background:
			newOverride = color.ParentBackground
		}

		if override == newOverride {
			return
		}

		newAnchor := strings.Replace(match[terminal.ANCHOR], string(override), string(newOverride), 1)
		e.previousActiveSegment.TrailingDiamond = strings.Replace(e.previousActiveSegment.TrailingDiamond, anchor, newAnchor, 1)
	}

	if len(match[terminal.BG]) > 0 {
		adjustOverride(match[terminal.ANCHOR], color.Ansi(match[terminal.BG]))
	}

	if len(match[terminal.FG]) > 0 {
		adjustOverride(match[terminal.ANCHOR], color.Ansi(match[terminal.FG]))
	}
}

func (e *Engine) rectifyTerminalWidth(diff int) {
	// Since the terminal width may not be given by the CLI flag, we should always call this here.
	_, err := e.Env.TerminalWidth()
	if err != nil {
		// Skip when we're unable to determine the terminal width.
		return
	}

	e.Env.Flags().TerminalWidth += diff
}

// New returns a prompt engine initialized with the
// given configuration options, and is ready to print any
// of the prompt components.
func New(flags *runtime.Flags) *Engine {
	env := &runtime.Terminal{}
	env.Init(flags)

	_, reload := cache.Get[bool](cache.Session, config.RELOAD)
	cfg := config.Get(flags.ConfigPath, reload)

	template.Init(env, cfg.Var, cfg.Maps)

	flags.HasExtra = cfg.DebugPrompt != nil ||
		cfg.SecondaryPrompt != nil ||
		cfg.TransientPrompt != nil ||
		cfg.ValidLine != nil ||
		cfg.ErrorLine != nil

	// when we print using https://github.com/akinomyoga/ble.sh, this needs to be unescaped for certain prompts
	sh := env.Shell()
	if sh == shell.BASH && !flags.Escape {
		sh = shell.GENERIC
	}

	terminal.Init(sh)
	terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	terminal.Colors = cfg.MakeColors(env)
	terminal.Plain = flags.Plain

	eng := &Engine{
		Config:      cfg,
		Env:         env,
		Plain:       flags.Plain,
		forceRender: flags.Force || len(env.Getenv("POSH_FORCE_RENDER")) > 0,
		prompt:      strings.Builder{},
	}

	// Pre-allocate prompt builder capacity to reduce allocations during rendering
	eng.prompt.Grow(512) // Start with 512 bytes capacity, will grow as needed

	switch env.Shell() {
	case shell.XONSH:
		// In Xonsh, the behavior of wrapping at the end of a prompt line is inconsistent across different operating systems.
		// On Windows, it wraps before the last cell on the terminal screen, that is, the last cell is never available for a prompt line.
		if env.GOOS() == runtime.WINDOWS {
			eng.rectifyTerminalWidth(-1)
		}
	case shell.ELVISH:
		// In Elvish, the case is similar to that in Xonsh.
		// However, on Windows, we have to reduce the terminal width by 1 again to ensure that newlines are displayed correctly.
		diff := -1
		if env.GOOS() == runtime.WINDOWS {
			diff = -2
		}
		eng.rectifyTerminalWidth(diff)
	case shell.PWSH, shell.PWSH5:
		// when in PowerShell, and force patching the bleed bug
		// we need to reduce the terminal width by 1 so the last
		// character isn't cut off by the ANSI escape sequences
		// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
		if cfg.PatchPwshBleed {
			eng.rectifyTerminalWidth(-1)
		}
	}

	return eng
}
