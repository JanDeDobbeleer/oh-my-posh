package prompt

import (
	"strings"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

var cycle *color.Cycle = &color.Cycle{}

type Engine struct {
	Env                   runtime.Environment
	streamingResults      chan *config.Segment
	abort                 chan struct{}
	done                  chan struct{}
	Config                *config.Config
	activeSegment         *config.Segment
	previousActiveSegment *config.Segment
	// blockTailColors is the last rendered segment's colors, captured just
	// before previousActiveSegment resets to nil at the end of a block. A
	// block's own Filler (see shouldFill) renders after terminal.String()
	// has already cleared the parent stack for the next block, so a filler
	// template using <parentBackground>/<parentForeground> needs this to
	// resolve against the block it is padding rather than an empty stack.
	blockTailColors   *color.Set
	pendingSegments   sync.Map
	rprompt           string
	Overflow          config.Overflow
	prompt            strings.Builder
	allBlocks         []*config.Block
	currentLineLength int
	Padding           int
	rpromptLength     int
	Plain             bool
	forceRender       bool
}

const (
	PRIMARY         = "primary"
	TRANSIENT       = "transient"
	TRANSIENT_RIGHT = "transient-right"
	DEBUG           = "debug"
	SECONDARY       = "secondary"
	RIGHT           = "right"
	TOOLTIP         = "tooltip"
	VALID           = "valid"
	ERROR           = "error"
	PREVIEW         = "preview"
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
	pwdType, err := template.RenderTrusted(e.Config.PWD, nil)
	if err != nil || pwdType == "" {
		return
	}

	// Convert to Windows path when in WSL
	if e.Env.IsWsl() {
		pwd = e.Env.ConvertToWindowsPath(pwd)
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
		log.Debug("no filler specified")
		return "", false
	}

	if padLength < 0 {
		log.Debug("padding length is negative")
		return "", false
	}

	e.Padding = padLength

	defer func() {
		e.Padding = 0
	}()

	var err error
	if filler, err = template.RenderTrusted(filler, e); err != nil {
		return "", false
	}

	// allow for easy color overrides and templates
	terminal.SetColors("default", "default")

	// the block's own segments already reset the parent stack when their
	// terminal.String() call produced blockText - reseed the one entry a
	// <parentBackground>/<parentForeground> anchor in the filler needs.
	if e.blockTailColors != nil {
		terminal.ParentColors = append(terminal.ParentColors, e.blockTailColors)
	}

	terminal.Write("", "", filler)
	filler, lenFiller := terminal.String()
	if lenFiller == 0 {
		log.Debug("filler has no length")
		return "", false
	}

	repeat := padLength / lenFiller
	unfilled := padLength % lenFiller
	txt := strings.Repeat(filler, repeat) + strings.Repeat(" ", unfilled)
	log.Debug("filling with", txt)
	return txt, true
}

func (e *Engine) getTitleTemplateText() string {
	if txt, err := template.RenderTrusted(e.Config.ConsoleTitleTemplate, nil); err == nil {
		return txt
	}

	return ""
}

// renderLaunchedBlock renders a block using pre-collected segment results
// (see drainBlockResults). executed must be fully populated for every block
// in the prompt before this is called so that cross-block .Segments.X
// dependencies resolve in both directions.
func (e *Engine) renderLaunchedBlock(block *config.Block, results []*config.Segment, executed map[string]bool, cancelNewline bool) bool {
	var blockText string
	var length int

	if results != nil {
		blockText, length = e.renderBlockSegments(results, block, executed)
	}

	// do not print anything when we don't have any text unless forced
	if !block.Force && length == 0 {
		return false
	}

	return e.writeBlock(block, blockText, length, cancelNewline)
}

// writeBlock handles the common logic for writing a block to the prompt
func (e *Engine) writeBlock(block *config.Block, blockText string, length int, cancelNewline bool) bool {
	defer func() {
		e.applyPowerShellBleedPatch()
	}()

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
			e.write(blockText)
			return true
		}

		if block.Alignment != config.Right {
			return false
		}

		space, OK := e.canWriteRightBlock(length, false)

		// we can't print the right block as there's not enough room available
		if !OK {
			e.Overflow = block.Overflow

			switch e.Overflow {
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
			e.write(blockText)
			return true
		}

		if space > 0 {
			e.write(strings.Repeat(" ", space))
		}

		e.write(blockText)
	case config.RPrompt:
		e.rprompt = blockText
		e.rpromptLength = length
	}

	return true
}

// renderBlockFromCache re-renders a block using existing segment data without re-execution
func (e *Engine) renderBlockFromCache(block *config.Block, cancelNewline bool) bool {
	if block.RestartCycle {
		cycle = &e.Config.Cycle
	}

	// Re-render all segments in the block
	for segmentIndex, segment := range block.Segments {
		// Allow pending segments to render (they show "..." text)
		if !segment.Pending && !segment.Enabled && segment.ResolveStyle() != config.Accordion {
			continue
		}

		// Render segment text (will use pending state if still pending)
		if !segment.Render(segmentIndex, e.forceRender) {
			continue
		}

		if colors, newCycle := cycle.Loop(); colors != nil {
			cycle = &newCycle
			segment.Foreground = colors.Foreground
			segment.Background = colors.Background
		}

		if terminal.Len() == 0 && len(block.LeadingDiamond) > 0 {
			segment.LeadingDiamond = block.LeadingDiamond
		}

		e.setActiveSegment(segment)
		e.renderActiveSegment()
	}

	if e.activeSegment != nil && len(block.TrailingDiamond) > 0 {
		e.activeSegment.TrailingDiamond = block.TrailingDiamond
	}

	e.writeSeparator(true)
	e.activeSegment = nil
	e.captureBlockTailColors()
	e.previousActiveSegment = nil

	blockText, length := terminal.String()

	// do not print anything when we don't have any text unless forced
	if !block.Force && length == 0 {
		return false
	}

	return e.writeBlock(block, blockText, length, cancelNewline)
}

func (e *Engine) applyPowerShellBleedPatch() {
	// when in PowerShell, we need to clear the line after the prompt
	// to avoid the background being printed on the next line
	// when at the end of the buffer.
	// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
	if e.Env.Shell() != shell.PWSH {
		return
	}

	// only do this when enabled
	if !e.Config.PatchPwshBleed {
		return
	}

	e.write(terminal.ClearAfter())
}

// minGradientCellsPerStop is the minimum number of visible cells a gradient stop needs to
// render as a smooth blend rather than a discrete color block; see Amendment 3 in the
// gradient spec. Below cells < minGradientCellsPerStop*stops, collapseGradient replaces the
// whole channel with the gradient's last stop.
const minGradientCellsPerStop = 2

func (e *Engine) setActiveSegment(segment *config.Segment) {
	e.activeSegment = segment
	terminal.Interactive = segment.Interactive

	background := resolvePaletteReference(segment.ResolveBackground())
	foreground := resolvePaletteReference(segment.ResolveForeground())

	// palette-resolved values are written back to the segment's resolved-color cache
	// so a palette entry holding a gradient is visible to every downstream consumer
	// (separators, diamonds, parent color references), not just this render call.
	if background != segment.ResolveBackground() {
		segment.CollapseBackground(background)
	}

	if foreground != segment.ResolveForeground() {
		segment.CollapseForeground(foreground)
	}

	// the collapse decision is made once per segment, before anything renders, so every
	// consumer of the segment's resolved colors (this Write call, separators, diamonds,
	// parent color references) agrees on the same solid color; see collapseGradient.
	// Pending placeholders are exempt: they are transient and should preview the
	// segment's gradient rather than flash a collapsed solid color mid-stream.
	if !segment.Pending && (background.IsGradient() || foreground.IsGradient()) {
		cells := terminal.VisibleCells(segment.Text())

		if collapsed, ok := collapseGradient(background, cells); ok {
			background = collapsed
			segment.CollapseBackground(background)
		}

		if collapsed, ok := collapseGradient(foreground, cells); ok {
			foreground = collapsed
			segment.CollapseForeground(foreground)
		}
	}

	terminal.SetColors(background, foreground)
}

// resolvePaletteReference expands a palette reference (p:name) so a palette entry
// holding a gradient is visible to the engine's gradient handling; without this,
// IsGradient/GradientLast run on the literal "p:name" string and every gradient
// rule is silently skipped for palette-referenced gradients.
func resolvePaletteReference(c color.Ansi) color.Ansi {
	if terminal.Colors == nil {
		return c
	}

	resolved, err := terminal.Colors.Resolve(c)
	if err != nil {
		return c
	}

	return resolved
}

// collapseGradient reports whether c must collapse to a single solid color because the
// segment has fewer than minGradientCellsPerStop visible cells per stop, returning that
// color (the gradient's last stop) when so. A non-gradient value, or a syntactically invalid
// gradient (nil GradientStops), is left untouched: the writer's existing per-call fallback
// handles those.
func collapseGradient(c color.Ansi, cells int) (color.Ansi, bool) {
	if !c.IsGradient() {
		return c, false
	}

	stops := c.GradientStops()
	if len(stops) < 2 {
		return c, false
	}

	if cells >= minGradientCellsPerStop*len(stops) {
		return c, false
	}

	return stops[len(stops)-1], true
}

// backgroundEdge collapses a segment's background gradient to its last stop,
// resolving a keyword stop (foreground, background) against the SAME segment's
// colors so edge consumers never leak a keyword into the wrong context.
func backgroundEdge(segment *config.Segment) color.Ansi {
	background := resolvePaletteReference(segment.ResolveBackground())

	stop := background.GradientLast()

	switch stop { //nolint:exhaustive
	case color.Foreground:
		stop = resolvePaletteReference(segment.ResolveForeground()).GradientLast()
	case color.Background:
		// self-reference has no resolvable edge
		return color.Transparent
	}

	if stop == color.Foreground || stop == color.Background {
		return color.Transparent
	}

	return stop
}

func (e *Engine) renderActiveSegment() {
	e.writeSeparator(false)

	switch e.activeSegment.ResolveStyle() {
	case config.Plain, config.Powerline:
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Diamond:
		background := color.Transparent

		if e.previousActiveSegment != nil && e.previousActiveSegment.HasEmptyDiamondAtEnd() {
			// this is the previous segment's right edge; a gradient must show its last stop.
			background = backgroundEdge(e.previousActiveSegment)
		}

		terminal.Write(background, color.Background, e.activeSegment.LeadingDiamond)
		terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Accordion:
		// Render accordion segments if enabled OR pending (pending shows "..." text)
		if e.activeSegment.Enabled || e.activeSegment.Pending {
			terminal.Write(color.Background, color.Foreground, e.activeSegment.Text())
		}
	}

	e.previousActiveSegment = e.activeSegment

	terminal.SetParentColors(e.previousActiveSegment.ResolveBackground(), e.previousActiveSegment.ResolveForeground())
}

// captureBlockTailColors snapshots the last rendered segment's colors into
// blockTailColors just before previousActiveSegment resets to nil, so a
// later shouldFill call for this same block's Filler can still resolve
// <parentBackground>/<parentForeground> after terminal.String() has already
// cleared the parent stack for the next block.
func (e *Engine) captureBlockTailColors() {
	if e.previousActiveSegment == nil {
		e.blockTailColors = nil
		return
	}

	e.blockTailColors = &color.Set{
		Background: e.previousActiveSegment.ResolveBackground(),
		Foreground: e.previousActiveSegment.ResolveForeground(),
	}
}

func (e *Engine) writeSeparator(final bool) {
	if e.activeSegment == nil {
		return
	}

	isCurrentDiamond := e.activeSegment.ResolveStyle() == config.Diamond
	if final && isCurrentDiamond {
		// the trailing diamond sits at the segment's right edge; a gradient
		// background must render as its last stop, not the writer's cells==1 default.
		diamondColor := color.Background
		if resolvePaletteReference(e.activeSegment.ResolveBackground()).IsGradient() {
			diamondColor = backgroundEdge(e.activeSegment)
		}

		terminal.Write(color.Transparent, diamondColor, e.resolveTrailingDiamond())
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

// getPowerlineColor resolves the separator symbol's color, which always sits at the
// previous segment's right edge; a gradient background must collapse to its last stop,
// resolved against the previous segment's own context (see backgroundEdge).
func (e *Engine) getPowerlineColor() color.Ansi {
	if e.previousActiveSegment == nil {
		return color.Transparent
	}

	if e.previousActiveSegment.ResolveStyle() == config.Diamond && e.previousActiveSegment.TrailingDiamond == "" {
		return backgroundEdge(e.previousActiveSegment)
	}

	if e.activeSegment.ResolveStyle() == config.Diamond && e.activeSegment.LeadingDiamond == "" {
		return backgroundEdge(e.previousActiveSegment)
	}

	if !e.previousActiveSegment.IsPowerline() {
		return color.Transparent
	}

	return backgroundEdge(e.previousActiveSegment)
}

// resolveTrailingDiamond rewrites a `background` keyword inside the active segment's
// trailing diamond template to the gradient's resolved last stop. The diamond renders
// in its own Write with no gradient cell context, so the keyword would otherwise
// collapse to the FIRST stop — a visible seam at the segment's right edge.
func (e *Engine) resolveTrailingDiamond() string {
	diamond := e.activeSegment.TrailingDiamond

	if !strings.Contains(diamond, string(color.Background)) {
		return diamond
	}

	if !resolvePaletteReference(e.activeSegment.ResolveBackground()).IsGradient() {
		return diamond
	}

	match := regex.FindNamedRegexMatch(terminal.AnchorRegex, diamond)
	if len(match) == 0 {
		return diamond
	}

	edge := backgroundEdge(e.activeSegment)
	if edge.IsClear() {
		return diamond
	}

	anchor := match[terminal.ANCHOR]
	adjusted := strings.ReplaceAll(anchor, string(color.Background), edge.String())

	return strings.Replace(diamond, anchor, adjusted, 1)
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

func (e *Engine) cancelNewline() bool {
	row, _ := e.Env.CursorPosition()
	return e.Env.Flags().Cleared || e.Env.Flags().PromptCount == 1 || row == 1
}

// New returns a prompt engine initialized with the
// given configuration options, and is ready to print any
// of the prompt components.
func New(flags *runtime.Flags) *Engine {
	env := &runtime.Terminal{}
	env.Init(flags)

	reload, _ := cache.Get[bool](cache.Device, config.RELOAD)
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
	case shell.PWSH:
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
