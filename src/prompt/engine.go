package prompt

import (
	"encoding/json"
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

type promptCache struct {
	Prompt            string
	CurrentLineLength int
	RPrompt           string
	RPromptLength     int
}

type Engine struct {
	Config *config.Config
	Env    runtime.Environment
	Plain  bool

	prompt            strings.Builder
	currentLineLength int
	rprompt           string
	rpromptLength     int

	activeSegment         *config.Segment
	previousActiveSegment *config.Segment

	promptCache *promptCache
}

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

	return availableSpace, canWrite
}

func (e *Engine) pwd() {
	// only print when supported
	sh := e.Env.Shell()
	if sh == shell.ELVISH || sh == shell.XONSH {
		return
	}
	// only print when relevant
	if len(e.Config.PWD) == 0 && !e.Config.OSC99 {
		return
	}

	cwd := e.Env.Pwd()

	// Backwards compatibility for deprecated OSC99
	if e.Config.OSC99 {
		e.write(terminal.Pwd(terminal.OSC99, "", "", cwd))
		return
	}

	// Allow template logic to define when to enable the PWD (when supported)
	tmpl := &template.Text{
		Template: e.Config.PWD,
		Env:      e.Env,
	}

	pwdType, err := tmpl.Render()
	if err != nil || len(pwdType) == 0 {
		return
	}

	user := e.Env.User()
	host, _ := e.Env.Host()
	e.write(terminal.Pwd(pwdType, user, host, cwd))
}

func (e *Engine) getNewline() string {
	// WARP terminal will remove \n from the prompt, so we hack a newline in
	if e.isWarp() {
		return terminal.LineBreak()
	}

	// TCSH needs a space before the LITERAL newline character or it will not render correctly
	// don't ask why, it be like that sometimes.
	// https://unix.stackexchange.com/questions/99101/properly-defining-a-multi-line-prompt-in-tcsh#comment1342462_322189
	if e.Env.Shell() == shell.TCSH {
		return ` \n`
	}

	return "\n"
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

	if padLength <= 0 {
		return "", false
	}

	// allow for easy color overrides and templates
	terminal.Write("", "", filler)
	filler, lenFiller := terminal.String()
	if lenFiller == 0 {
		return "", false
	}

	repeat := padLength / lenFiller
	return strings.Repeat(filler, repeat), true
}

func (e *Engine) getTitleTemplateText() string {
	tmpl := &template.Text{
		Template: e.Config.ConsoleTitleTemplate,
		Env:      e.Env,
	}
	if text, err := tmpl.Render(); err == nil {
		return text
	}
	return ""
}

func (e *Engine) renderBlock(block *config.Block, cancelNewline bool) bool {
	defer e.patchPowerShellBleed()

	// This is deprecated but we leave it in to not break configs
	// It is encouraged to use "newline": true on block level
	// rather than the standalone linebreak block
	if block.Type == config.LineBreak {
		// do not print a newline to avoid a leading space
		// when we're printing the first primary prompt in
		// the shell
		if !cancelNewline {
			e.writeNewline()
		}
		return false
	}

	block.Init(e.Env)

	if !block.Enabled() {
		return false
	}

	// do not print a newline to avoid a leading space
	// when we're printing the first primary prompt in
	// the shell
	if block.Newline && !cancelNewline {
		e.writeNewline()
	}

	text, length := e.renderBlockSegments(block)

	// do not print anything when we don't have any text
	if length == 0 {
		return false
	}

	switch block.Type { //nolint:exhaustive
	case config.Prompt:
		if block.VerticalOffset != 0 {
			e.write(terminal.ChangeLine(block.VerticalOffset))
		}

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

func (e *Engine) patchPowerShellBleed() {
	// when in PowerShell, we need to clear the line after the prompt
	// to avoid the background being printed on the next line
	// when at the end of the buffer.
	// See https://githue.com/JanDeDobbeleer/oh-my-posh/issues/65
	if e.Env.Shell() != shell.PWSH && e.Env.Shell() != shell.PWSH5 {
		return
	}

	// only do this when enabled
	if !e.Config.PatchPwshBleed {
		return
	}

	e.write(terminal.ClearAfter())
}

func (e *Engine) renderBlockSegments(block *config.Block) (string, int) {
	e.filterSegments(block)

	for i, segment := range block.Segments {
		if colors, newCycle := cycle.Loop(); colors != nil {
			cycle = &newCycle
			segment.Foreground = colors.Foreground
			segment.Background = colors.Background
		}

		if i == 0 && len(block.LeadingDiamond) > 0 {
			segment.LeadingDiamond = block.LeadingDiamond
		}

		if i == len(block.Segments)-1 && len(block.TrailingDiamond) > 0 {
			segment.TrailingDiamond = block.TrailingDiamond
		}

		e.setActiveSegment(segment)
		e.renderActiveSegment()
	}

	e.writeSeparator(true)

	e.activeSegment = nil
	e.previousActiveSegment = nil

	return terminal.String()
}

func (e *Engine) filterSegments(block *config.Block) {
	segments := make([]*config.Segment, 0)

	for _, segment := range block.Segments {
		if !segment.Enabled && segment.ResolveStyle() != config.Accordion {
			continue
		}

		segments = append(segments, segment)
	}

	block.Segments = segments
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
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text)
	case config.Diamond:
		background := color.Transparent

		if e.previousActiveSegment != nil && e.previousActiveSegment.HasEmptyDiamondAtEnd() {
			background = e.previousActiveSegment.ResolveBackground()
		}

		terminal.Write(background, color.Background, e.activeSegment.LeadingDiamond)
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text)
	case config.Accordion:
		if e.activeSegment.Enabled {
			terminal.Write(color.Background, color.Foreground, e.activeSegment.Text)
		}
	}

	e.previousActiveSegment = e.activeSegment

	terminal.SetParentColors(e.previousActiveSegment.ResolveBackground(), e.previousActiveSegment.ResolveForeground())
}

func (e *Engine) writeSeparator(final bool) {
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

	if e.activeSegment.InvertPowerline {
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

func (e *Engine) checkPromptCache() bool {
	data, ok := e.Env.Cache().Get(cache.PROMPTCACHE)
	if !ok {
		return false
	}

	e.promptCache = &promptCache{}
	err := json.Unmarshal([]byte(data), e.promptCache)
	if err != nil {
		return false
	}

	e.write(e.promptCache.Prompt)
	e.currentLineLength = e.promptCache.CurrentLineLength
	e.rprompt = e.promptCache.RPrompt
	e.rpromptLength = e.promptCache.RPromptLength

	return true
}

func (e *Engine) updatePromptCache(value *promptCache) {
	cacheJSON, err := json.Marshal(value)
	if err != nil {
		return
	}
	e.Env.Cache().Set(cache.PROMPTCACHE, string(cacheJSON), 1440)
}

// New returns a prompt engine initialized with the
// given configuration options, and is ready to print any
// of the prompt components.
func New(flags *runtime.Flags) *Engine {
	env := &runtime.Terminal{
		CmdFlags: flags,
	}

	env.Init()
	cfg := config.Load(env)

	if cfg.PatchPwshBleed {
		patchPowerShellBleed(env.Shell(), flags)
	}

	env.Var = cfg.Var
	flags.HasTransient = cfg.TransientPrompt != nil

	terminal.Init(env.Shell())
	terminal.BackgroundColor = shell.ConsoleBackgroundColor(env, cfg.TerminalBackground)
	terminal.Colors = cfg.MakeColors()
	terminal.Plain = flags.Plain

	eng := &Engine{
		Config: cfg,
		Env:    env,
		Plain:  flags.Plain,
	}

	return eng
}

func patchPowerShellBleed(sh string, flags *runtime.Flags) {
	// when in PowerShell, and force patching the bleed bug
	// we need to reduce the terminal width by 1 so the last
	// character isn't cut off by the ANSI escape sequences
	// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
	if sh != shell.PWSH && sh != shell.PWSH5 {
		return
	}

	// only do this when relevant
	if flags.TerminalWidth <= 0 {
		return
	}

	flags.TerminalWidth--
}
