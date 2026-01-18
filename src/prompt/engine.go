package prompt

import (
	"fmt"
	"strings"
	"sync"
	"time"

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

// TextCache is an interface for storing/retrieving cached segment text.
// The daemon provides an implementation that handles session management.
type TextCache interface {
	// Get retrieves cached text by key.
	Get(key string) (string, bool)
	// Set stores text with default strategy and TTL.
	Set(key string, value string)

	// GetWithAge retrieves cached text along with its age (time since creation).
	// Used for duration-based cache validation in daemon mode.
	GetWithAge(key string) (text string, age time.Duration, found bool)

	// SetWithConfig stores text with explicit cache configuration.
	// The segment's Cache config determines strategy and duration.
	SetWithConfig(key string, value string, cacheConfig *config.Cache)

	// ShouldRecompute determines if a segment should be recomputed based on its cache config.
	// Returns:
	//   - recompute: true if segment should be executed again
	//   - useCacheForPending: true if cached value should be shown during pending render
	ShouldRecompute(key string, cacheConfig *config.Cache) (recompute bool, useCacheForPending bool)
}

type Engine struct {
	Env                   runtime.Environment
	Writer                *terminal.Writer
	TextCache             TextCache
	pendingSegments       map[string]bool
	Config                *config.Config
	activeSegment         *config.Segment
	previousActiveSegment *config.Segment
	TemplateCache         *cache.Template
	cachedValues          map[string]string
	rprompt               string
	Overflow              config.Overflow
	prompt                strings.Builder
	currentLineLength     int
	Padding               int
	rpromptLength         int
	streamingMu           sync.Mutex
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

// PrimaryStreaming renders the primary prompt using parallel segment execution.
// It returns partial results after timeout, then continues updating in background.
// Used by the daemon.
func (e *Engine) PrimaryStreaming(timeout time.Duration, updateCallback func(string)) string {
	// Initialize streaming state
	e.pendingSegments = make(map[string]bool)
	e.cachedValues = make(map[string]string)
	var segmentsToExecute []*config.Segment
	var segmentsSkipped []*config.Segment // Segments with fresh cache that don't need execution
	var wg sync.WaitGroup

	// Collect segments and determine which need execution
	for _, block := range e.Config.Blocks {
		for _, segment := range block.Segments {
			// Pre-initialize writer and name for cache key generation
			_ = segment.MapSegmentWithWriter(e.Env)
			_ = segment.Name()

			// Check if we should recompute this segment (daemon mode cache logic)
			if e.TextCache != nil {
				cacheKey := segment.DaemonCacheKey()
				shouldRecompute, useCacheForPending := e.TextCache.ShouldRecompute(cacheKey, segment.Cache)

				if !shouldRecompute {
					// Cache is fresh: use cached value directly, skip execution
					if cachedText, ok := e.TextCache.Get(cacheKey); ok {
						segment.Enabled = true
						segment.SetText(cachedText)
						segmentsSkipped = append(segmentsSkipped, segment)
						continue
					}
				}

				// Store cached value for pending display if available
				if useCacheForPending {
					if cachedText, ok := e.TextCache.Get(cacheKey); ok {
						e.cachedValues[segment.Name()] = cachedText
					}
				}
			}

			segmentsToExecute = append(segmentsToExecute, segment)
		}
	}

	// Helper to start segment execution
	execute := func(segment *config.Segment, index int, out chan<- result) {
		// Execute segment (state modified in place on cloned segment)
		segment.Execute(e.Env)
		out <- result{segment, index}
	}

	// Start all segments that need execution in parallel
	results := make(chan result, len(segmentsToExecute))
	for i, segment := range segmentsToExecute {
		wg.Add(1)
		go func(s *config.Segment, idx int) {
			defer wg.Done()
			execute(s, idx, results)
		}(segment, i)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Wait for results or timeout
	timeoutChan := time.After(timeout)

	// Consume results until all done or timeout
	executedCount := 0
	totalCount := len(segmentsToExecute)

	// Map to track completed segments for the initial render pass
	completedSegments := make(map[*config.Segment]bool)

	// Mark skipped segments as already completed
	for _, s := range segmentsSkipped {
		completedSegments[s] = true
	}

	// Wait loop
	loop := true
	for loop {
		select {
		case res, ok := <-results:
			if !ok {
				loop = false
				break
			}
			executedCount++
			completedSegments[res.segment] = true
			if executedCount == totalCount {
				loop = false
			}
		case <-timeoutChan:
			loop = false
		}
	}

	// Identify pending segments (those that started but haven't finished)
	e.streamingMu.Lock()
	for _, s := range segmentsToExecute {
		if !completedSegments[s] {
			e.pendingSegments[s.Name()] = true
		}
	}
	e.streamingMu.Unlock()

	// Initial render
	initialPrompt := e.renderStreamingPrompt()

	// If there are pending segments, start background listener for updates
	if len(e.pendingSegments) > 0 {
		go func() {
			// Continue consuming results
			for res := range results {
				e.streamingMu.Lock()
				delete(e.pendingSegments, res.segment.Name())
				e.streamingMu.Unlock()

				// Notify daemon
				updateCallback(res.segment.Name())
			}
		}()
	}

	return initialPrompt
}

// ReRender re-renders the prompt using the current state of all segments.
// Called by the daemon when a background segment completes.
func (e *Engine) ReRender() string {
	e.streamingMu.Lock()
	defer e.streamingMu.Unlock()
	return e.renderStreamingPrompt()
}

// StreamingRPrompt returns the rprompt text from the last streaming render.
// Unlike RPrompt(), this does NOT re-execute segments — it returns the value
// computed during renderStreamingPrompt (which handles pending segments).
func (e *Engine) StreamingRPrompt() string {
	return e.rprompt
}

// renderStreamingPrompt renders the prompt handling pending vs completed segments.
func (e *Engine) renderStreamingPrompt() string {
	e.prompt.Reset()
	e.currentLineLength = 0
	e.rprompt = ""
	e.rpromptLength = 0

	needsPrimaryRightPrompt := e.needsPrimaryRightPrompt()
	e.writePrimaryPromptStreaming(needsPrimaryRightPrompt)

	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}
		// Warp doesn't support RPROMPT so we need to write it manually
		if e.isWarp() {
			e.writePrimaryRightPrompt()
			prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(e.string()))
			return prompt
		}

		prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(e.string()))
		prompt += fmt.Sprintf("\nRPROMPT=%s", shell.QuotePosixStr(e.rprompt))
		return prompt
	default:
		if !needsPrimaryRightPrompt {
			break
		}
		e.writePrimaryRightPrompt()
	}

	return e.string()
}

// writePrimaryPromptStreaming is a copy of writePrimaryPrompt but handles pending segments
func (e *Engine) writePrimaryPromptStreaming(needsPrimaryRPrompt bool) {
	if e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(terminal.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(terminal.PromptStart())
	}

	cycle = &e.Config.Cycle
	var cancelNewline, didRender bool

	for i, block := range e.Config.Blocks {
		if i == 0 {
			row, _ := e.Env.CursorPosition()
			cancelNewline = e.Env.Flags().Cleared || e.Env.Flags().PromptCount == 1 || row == 1
		}

		if i != 0 {
			cancelNewline = !didRender
		}

		if block.Type == config.RPrompt && !needsPrimaryRPrompt {
			continue
		}

		// Custom renderBlock logic for streaming
		if e.renderBlockStreaming(block, cancelNewline) {
			didRender = true
		}
	}

	if len(e.Config.ConsoleTitleTemplate) > 0 && !e.Env.Flags().Plain {
		title := e.getTitleTemplateText()
		e.write(terminal.FormatTitle(title))
	}

	if e.Config.FinalSpace {
		e.write(" ")
		e.currentLineLength++
	}

	if e.Config.ITermFeatures != nil && e.isIterm() {
		host, _ := e.Env.Host()
		e.write(terminal.RenderItermFeatures(e.Config.ITermFeatures, e.Env.Shell(), e.Env.Pwd(), e.Env.User(), host))
	}
}

// renderBlockStreaming renders a block, handling pending segments
func (e *Engine) renderBlockStreaming(block *config.Block, cancelNewline bool) bool {
	blockText, length := e.writeBlockSegmentsStreaming(block)
	return e.renderBlockWithText(block, blockText, length, cancelNewline)
}

func (e *Engine) writeBlockSegmentsStreaming(block *config.Block) (string, int) {
	segmentIndex := 0

	for _, segment := range block.Segments {
		// Check if pending
		isPending := e.pendingSegments[segment.Name()]

		if isPending {
			// Render as pending
			cachedVal := e.cachedValues[segment.Name()]
			enabled, text, background := segment.GetPendingText(cachedVal, e.Config)

			if !enabled {
				continue
			}

			// Store original state to restore later
			originalText := segment.Text()
			originalBackground := segment.Background
			originalBackgroundTemplates := segment.BackgroundTemplates
			originalEnabled := segment.Enabled

			// Update segment for pending render
			segment.SetText(text)
			if background != "" {
				segment.Background = background
			} else {
				switch {
				case segment.RenderPendingBackground != "":
					segment.Background = segment.RenderPendingBackground
				case e.Config.RenderPendingBackground != "":
					segment.Background = e.Config.RenderPendingBackground
				default:
					segment.Background = "darkGray"
				}
			}
			// Clear templates so they don't override the pending background
			segment.BackgroundTemplates = nil
			segment.Enabled = true

			// Render directly (bypass writeSegment to skip cycle color override)
			e.setActiveSegment(segment)
			e.renderActiveSegment()

			// Restore state
			segment.SetText(originalText)
			segment.Background = originalBackground
			segment.BackgroundTemplates = originalBackgroundTemplates
			segment.Enabled = originalEnabled
		} else {
			// Normal rendering
			// Segment.Execute has already run (or is finished)
			if segment.Render(segmentIndex, e.forceRender) {
				segmentIndex++
			}
			e.writeSegment(block, segment)
		}
	}

	if e.activeSegment != nil && len(block.TrailingDiamond) > 0 {
		e.activeSegment.TrailingDiamond = block.TrailingDiamond
	}

	e.writeSeparator(true)

	e.activeSegment = nil
	e.previousActiveSegment = nil

	return e.Writer.String()
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

	e.Padding = padLength

	defer func() {
		e.Padding = 0
	}()

	var err error
	if filler, err = template.Render(filler, e); err != nil {
		return "", false
	}

	// allow for easy color overrides and templates
	e.Writer.SetColors("default", "default")
	e.Writer.Write("", "", filler)
	filler, lenFiller := e.Writer.String()
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
	if txt, err := template.Render(e.Config.ConsoleTitleTemplate, nil); err == nil {
		return txt
	}

	return ""
}

// CacheSegmentText stores a segment's rendered text in TextCache.
// Called when a segment completes successfully in daemon mode.
// Uses DaemonCacheKey and stores with the segment's cache configuration.
func (e *Engine) CacheSegmentText(segment *config.Segment) {
	if e.TextCache == nil {
		return
	}

	if !segment.Enabled {
		return
	}

	key := segment.DaemonCacheKey()
	text := segment.Text()
	e.TextCache.SetWithConfig(key, text, segment.Cache)
}

// StreamingMu returns the mutex protecting streaming state.
func (e *Engine) StreamingMu() *sync.Mutex {
	return &e.streamingMu
}

// PendingSegments returns the set of currently pending segments.
func (e *Engine) PendingSegments() map[string]bool {
	return e.pendingSegments
}

func (e *Engine) renderBlock(block *config.Block, cancelNewline bool) bool {
	blockText, length := e.writeBlockSegments(block)
	return e.renderBlockWithText(block, blockText, length, cancelNewline)
}

// renderBlockWithText renders a block with pre-computed text and length.
func (e *Engine) renderBlockWithText(block *config.Block, blockText string, length int, cancelNewline bool) bool {
	// do not print anything when we don't have any text unless forced
	if !block.Force && length == 0 {
		return false
	}

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

func (e *Engine) setActiveSegment(segment *config.Segment) {
	e.activeSegment = segment
	e.Writer.Interactive = segment.Interactive
	e.Writer.SetColors(segment.ResolveBackground(), segment.ResolveForeground())
}

func (e *Engine) renderActiveSegment() {
	e.writeSeparator(false)

	switch e.activeSegment.ResolveStyle() {
	case config.Plain, config.Powerline:
		e.Writer.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Diamond:
		background := color.Transparent

		if e.previousActiveSegment != nil && e.previousActiveSegment.HasEmptyDiamondAtEnd() {
			background = e.previousActiveSegment.ResolveBackground()
		}

		e.Writer.Write(background, color.Background, e.activeSegment.LeadingDiamond)
		e.Writer.Write(color.Background, color.Foreground, e.activeSegment.Text())
	case config.Accordion:
		if e.activeSegment.Enabled {
			e.Writer.Write(color.Background, color.Foreground, e.activeSegment.Text())
		}
	}

	e.previousActiveSegment = e.activeSegment

	e.Writer.SetParentColors(e.previousActiveSegment.ResolveBackground(), e.previousActiveSegment.ResolveForeground())
}

func (e *Engine) writeSeparator(final bool) {
	if e.activeSegment == nil {
		return
	}

	isCurrentDiamond := e.activeSegment.ResolveStyle() == config.Diamond
	if final && isCurrentDiamond {
		e.Writer.Write(color.Transparent, color.Background, e.activeSegment.TrailingDiamond)
		return
	}

	isPreviousDiamond := e.previousActiveSegment != nil && e.previousActiveSegment.ResolveStyle() == config.Diamond
	if isPreviousDiamond {
		e.adjustTrailingDiamondColorOverrides()
	}

	if isPreviousDiamond && isCurrentDiamond && e.activeSegment.LeadingDiamond == "" {
		e.Writer.Write(color.Background, color.ParentBackground, e.previousActiveSegment.TrailingDiamond)
		return
	}

	if isPreviousDiamond && len(e.previousActiveSegment.TrailingDiamond) > 0 {
		e.Writer.Write(color.Transparent, color.ParentBackground, e.previousActiveSegment.TrailingDiamond)
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
		e.Writer.Write(color.Transparent, color.Background, e.activeSegment.LeadingPowerlineSymbol)
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
		e.Writer.Write(e.getPowerlineColor(), bgColor, symbol)
		return
	}

	e.Writer.Write(bgColor, e.getPowerlineColor(), symbol)
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

// SaveTemplateCache persists the Engine's template cache to session storage.
// This allows secondary/transient prompts to access segment data from the primary prompt.
// Only used in CLI mode - daemon mode returns all prompts in a single response.
func (e *Engine) SaveTemplateCache() {
	if e.TemplateCache != nil {
		// Convert *cache.Template to whatever format template.SaveCache expects,
		// or update template.SaveCache to accept an argument.
		// In oh-my-posh-before-daemon, template.SaveCache() takes NO arguments and uses the global `Cache`.
		// But here we want to save e.TemplateCache.
		// So we might need to update template.SaveCache in oh-my-posh-before-daemon/src/template/cache.go
		// OR we can hack it by setting the global Cache to e.TemplateCache before calling SaveCache.

		// Let's modify template.SaveCache to be more flexible in a separate step if needed.
		// For now, let's just make this method exist and do nothing or do the hack.
		template.Cache = e.TemplateCache
		template.SaveCache()
	}
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

	writer := terminal.NewWriter(sh)
	writer.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	writer.Colors = cfg.MakeColors(env)
	writer.Plain = flags.Plain

	eng := &Engine{
		Config:        cfg,
		Env:           env,
		Writer:        writer,
		Plain:         flags.Plain,
		TemplateCache: template.NewCache(env, cfg.Var, cfg.Maps),
		forceRender:   flags.Force || len(env.Getenv("POSH_FORCE_RENDER")) > 0,
		prompt:        strings.Builder{},
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
