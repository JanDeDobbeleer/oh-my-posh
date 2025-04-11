package prompt

import (
	"strings"

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
	currentLineLength     int
	rpromptLength         int
	Padding               int
	Plain                 bool
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
)

func (e *Engine) write(text string) {
	e.prompt.WriteString(text)
}

func (e *Engine) string() string {
	text := e.prompt.String()
	e.prompt.Reset()
	return text
}

func (e *Engine) canWriteRightBlock(length int, rprompt bool) (int, bool) {
	if rprompt && (len(e.rprompt) == 0) {
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
	if len(e.Config.PWD) == 0 {
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
	tmpl := &template.Text{
		Template: e.Config.PWD,
	}

	pwdType, err := tmpl.Render()
	if err != nil || len(pwdType) == 0 {
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
	// For Elvish on Windows, we do this to prevent cutting off a right-aligned block.
	// For Tcsh, we do this to prevent a newline character from being printed.
	switch {
	case e.isWarp():
		fallthrough
	case e.Env.Shell() == shell.ELVISH && e.Env.GOOS() == runtime.WINDOWS:
		fallthrough
	case e.Env.Shell() == shell.TCSH:
		return terminal.LineBreak()
	default:
		return newline
	}
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
	if len(filler) == 0 {
		return "", false
	}

	tmpl := &template.Text{
		Template: filler,
		Context:  e,
	}

	var err error
	if filler, err = tmpl.Render(); err != nil {
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
	text := strings.Repeat(filler, repeat) + strings.Repeat(" ", unfilled)
	return text, true
}

func (e *Engine) getTitleTemplateText() string {
	tmpl := &template.Text{
		Template: e.Config.ConsoleTitleTemplate,
	}
	if text, err := tmpl.Render(); err == nil {
		return text
	}
	return ""
}

func (e *Engine) renderBlock(block *config.Block, cancelNewline bool) bool {
	text, length := e.writeBlockSegments(block)

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
			e.write(text)
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
				if padText, OK := e.shouldFill(block.Filler, space+length); OK {
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
			e.write(text)
			return true
		}

		var prompt string

		if space > 0 {
			prompt += strings.Repeat(" ", space)
		}

		prompt += text
		e.write(prompt)
	case config.RPrompt:
		e.rprompt = text
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

	if isPreviousDiamond && isCurrentDiamond && len(e.activeSegment.LeadingDiamond) == 0 {
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

		if isPowerline && len(e.activeSegment.LeadingPowerlineSymbol) == 0 {
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
	if len(symbol) == 0 {
		return
	}

	bgColor := color.Background
	if final || !isPowerline {
		bgColor = color.Transparent
	}

	if e.activeSegment.ResolveStyle() == config.Diamond && len(e.activeSegment.LeadingDiamond) == 0 {
		bgColor = color.Background
	}

	if e.activeSegment.InvertPowerline || e.previousActiveSegment.InvertPowerline {
		terminal.Write(e.getPowerlineColor(), bgColor, symbol)
		return
	}

	terminal.Write(bgColor, e.getPowerlineColor(), symbol)
}

func (e *Engine) getPowerlineColor() color.Ansi {
	if e.previousActiveSegment == nil {
		return color.Transparent
	}

	if e.previousActiveSegment.ResolveStyle() == config.Diamond && len(e.previousActiveSegment.TrailingDiamond) == 0 {
		return e.previousActiveSegment.ResolveBackground()
	}

	if e.activeSegment.ResolveStyle() == config.Diamond && len(e.activeSegment.LeadingDiamond) == 0 {
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
	if e.previousActiveSegment == nil || len(e.previousActiveSegment.TrailingDiamond) == 0 {
		return
	}

	if !strings.Contains(e.previousActiveSegment.TrailingDiamond, string(color.Background)) && !strings.Contains(e.previousActiveSegment.TrailingDiamond, string(color.Foreground)) {
		return
	}

	match := regex.FindNamedRegexMatch(terminal.AnchorRegex, e.previousActiveSegment.TrailingDiamond)
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
	flags.Config = config.Path(flags.Config)
	cfg := config.Load(flags.Config, flags.Shell, flags.Migrate)

	env := &runtime.Terminal{}
	env.Init(flags)

	template.Init(env, cfg.Var)

	flags.HasExtra = cfg.DebugPrompt != nil ||
		cfg.SecondaryPrompt != nil ||
		cfg.TransientPrompt != nil ||
		cfg.ValidLine != nil ||
		cfg.ErrorLine != nil

	terminal.Init(env.Shell())
	terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	terminal.Colors = cfg.MakeColors(env)
	terminal.Plain = flags.Plain

	eng := &Engine{
		Config: cfg,
		Env:    env,
		Plain:  flags.Plain,
	}

	switch env.Shell() {
	case shell.XONSH:
		// In Xonsh, the behavior of wrapping at the end of a prompt line is inconsistent across platforms.
		// On Windows, it wraps before the rightmost cell on the terminal screen, that is, the rightmost cell is never available for a prompt line.
		if eng.Env.GOOS() == runtime.WINDOWS {
			eng.rectifyTerminalWidth(-1)
		}
	case shell.TCSH, shell.ELVISH:
		// In Tcsh, newlines in a prompt are badly translated.
		// No silver bullet here. We have to reduce the terminal width by 1 so a right-aligned block will not be broken.
		// In Elvish, the behavior is similar to that in Xonsh, but we do this for all platforms.
		eng.rectifyTerminalWidth(-1)
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
