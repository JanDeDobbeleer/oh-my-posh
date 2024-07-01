package engine

import (
	"strings"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// BlockType type of block
type BlockType string

// BlockAlignment aligment of a Block
type BlockAlignment string

// Overflow defines how to handle a right block that overflows with the previous block
type Overflow string

const (
	// Prompt writes one or more Segments
	Prompt BlockType = "prompt"
	// LineBreak creates a line break in the prompt
	LineBreak BlockType = "newline"
	// RPrompt is a right aligned prompt
	RPrompt BlockType = "rprompt"
	// Left aligns left
	Left BlockAlignment = "left"
	// Right aligns right
	Right BlockAlignment = "right"
	// Break adds a line break
	Break Overflow = "break"
	// Hide hides the block
	Hide Overflow = "hide"
)

// Block defines a part of the prompt with optional segments
type Block struct {
	Type      BlockType      `json:"type,omitempty" toml:"type,omitempty"`
	Alignment BlockAlignment `json:"alignment,omitempty" toml:"alignment,omitempty"`
	Segments  []*Segment     `json:"segments,omitempty" toml:"segments,omitempty"`
	Newline   bool           `json:"newline,omitempty" toml:"newline,omitempty"`
	Filler    string         `json:"filler,omitempty" toml:"filler,omitempty"`
	Overflow  Overflow       `json:"overflow,omitempty" toml:"overflow,omitempty"`

	LeadingDiamond  string `json:"leading_diamond,omitempty" toml:"leading_diamond,omitempty"`
	TrailingDiamond string `json:"trailing_diamond,omitempty" toml:"trailing_diamond,omitempty"`

	// Deprecated: keep the logic for legacy purposes
	HorizontalOffset int `json:"horizontal_offset,omitempty" toml:"horizontal_offset,omitempty"`
	VerticalOffset   int `json:"vertical_offset,omitempty" toml:"vertical_offset,omitempty"`

	MaxWidth int `json:"max_width,omitempty" toml:"max_width,omitempty"`
	MinWidth int `json:"min_width,omitempty" toml:"min_width,omitempty"`

	env                   platform.Environment
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (b *Block) Init(env platform.Environment) {
	b.env = env
	b.executeSegmentLogic()
}

func (b *Block) InitPlain(env platform.Environment, config *Config) {
	terminal.Init(shell.GENERIC)
	terminal.BackgroundColor = shell.ConsoleBackgroundColor(env, config.TerminalBackground)
	terminal.AnsiColors = config.MakeColors()
	terminal.TrueColor = env.Flags().TrueColor

	b.env = env
	b.executeSegmentLogic()
}

func (b *Block) executeSegmentLogic() {
	if shouldHideForWidth(b.env, b.MinWidth, b.MaxWidth) {
		return
	}

	b.setEnabledSegments()
	b.setSegmentsText()
}

func (b *Block) setActiveSegment(segment *Segment) {
	b.activeSegment = segment
	terminal.Interactive = segment.Interactive
	terminal.SetColors(segment.background(), segment.foreground())
}

func (b *Block) Enabled() bool {
	if b.Type == LineBreak {
		return true
	}

	for _, segment := range b.Segments {
		if segment.Enabled {
			return true
		}
	}

	return false
}

func (b *Block) setEnabledSegments() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.Segments))
	defer wg.Wait()

	for _, segment := range b.Segments {
		go func(s *Segment) {
			defer wg.Done()
			s.SetEnabled(b.env)
		}(segment)
	}
}

func (b *Block) setSegmentsText() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.Segments))
	defer wg.Wait()
	for _, segment := range b.Segments {
		go func(s *Segment) {
			defer wg.Done()
			s.SetText()
		}(segment)
	}
}

func (b *Block) RenderSegments() (string, int) {
	b.filterSegments()

	for i, segment := range b.Segments {
		if colors, newCycle := cycle.Loop(); colors != nil {
			cycle = &newCycle
			segment.colors = colors
		}

		if i == 0 && len(b.LeadingDiamond) > 0 {
			segment.LeadingDiamond = b.LeadingDiamond
		}

		if i == len(b.Segments)-1 && len(b.TrailingDiamond) > 0 {
			segment.TrailingDiamond = b.TrailingDiamond
		}

		b.setActiveSegment(segment)
		b.renderActiveSegment()
	}

	b.writeSeparator(true)

	return terminal.String()
}

func (b *Block) filterSegments() {
	segments := make([]*Segment, 0)

	for _, segment := range b.Segments {
		if !segment.Enabled && segment.style() != Accordion {
			continue
		}

		segments = append(segments, segment)
	}

	b.Segments = segments
}

func (b *Block) renderActiveSegment() {
	b.writeSeparator(false)
	switch b.activeSegment.style() {
	case Plain, Powerline:
		terminal.Write(terminal.Background, terminal.Foreground, b.activeSegment.text)
	case Diamond:
		background := terminal.Transparent
		if b.previousActiveSegment != nil && b.previousActiveSegment.hasEmptyDiamondAtEnd() {
			background = b.previousActiveSegment.background()
		}
		terminal.Write(background, terminal.Background, b.activeSegment.LeadingDiamond)
		terminal.Write(terminal.Background, terminal.Foreground, b.activeSegment.text)
	case Accordion:
		if b.activeSegment.Enabled {
			terminal.Write(terminal.Background, terminal.Foreground, b.activeSegment.text)
		}
	}
	b.previousActiveSegment = b.activeSegment
	terminal.SetParentColors(b.previousActiveSegment.background(), b.previousActiveSegment.foreground())
}

func (b *Block) writeSeparator(final bool) {
	isCurrentDiamond := b.activeSegment.style() == Diamond
	if final && isCurrentDiamond {
		terminal.Write(terminal.Transparent, terminal.Background, b.activeSegment.TrailingDiamond)
		return
	}

	isPreviousDiamond := b.previousActiveSegment != nil && b.previousActiveSegment.style() == Diamond
	if isPreviousDiamond {
		b.adjustTrailingDiamondColorOverrides()
	}

	if isPreviousDiamond && isCurrentDiamond && len(b.activeSegment.LeadingDiamond) == 0 {
		terminal.Write(terminal.Background, terminal.ParentBackground, b.previousActiveSegment.TrailingDiamond)
		return
	}

	if isPreviousDiamond && len(b.previousActiveSegment.TrailingDiamond) > 0 {
		terminal.Write(terminal.Transparent, terminal.ParentBackground, b.previousActiveSegment.TrailingDiamond)
	}

	isPowerline := b.activeSegment.isPowerline()

	shouldOverridePowerlineLeadingSymbol := func() bool {
		if !isPowerline {
			return false
		}

		if isPowerline && len(b.activeSegment.LeadingPowerlineSymbol) == 0 {
			return false
		}

		if b.previousActiveSegment != nil && b.previousActiveSegment.isPowerline() {
			return false
		}

		return true
	}

	if shouldOverridePowerlineLeadingSymbol() {
		terminal.Write(terminal.Transparent, terminal.Background, b.activeSegment.LeadingPowerlineSymbol)
		return
	}

	resolvePowerlineSymbol := func() string {
		if isPowerline {
			return b.activeSegment.PowerlineSymbol
		}

		if b.previousActiveSegment != nil && b.previousActiveSegment.isPowerline() {
			return b.previousActiveSegment.PowerlineSymbol
		}

		return ""
	}

	symbol := resolvePowerlineSymbol()
	if len(symbol) == 0 {
		return
	}

	bgColor := terminal.Background
	if final || !isPowerline {
		bgColor = terminal.Transparent
	}

	if b.activeSegment.style() == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		bgColor = terminal.Background
	}

	if b.activeSegment.InvertPowerline {
		terminal.Write(b.getPowerlineColor(), bgColor, symbol)
		return
	}

	terminal.Write(bgColor, b.getPowerlineColor(), symbol)
}

func (b *Block) adjustTrailingDiamondColorOverrides() {
	// as we now already adjusted the activeSegment, we need to change the value
	// of background and foreground to parentBackground and parentForeground
	// this will still break when using parentBackground and parentForeground as keywords
	// in a trailing diamond, but let's fix that when it happens as it requires either a rewrite
	// of the logic for diamonds or storing grandparents as well like one happy family.
	if b.previousActiveSegment == nil || len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return
	}

	if !strings.Contains(b.previousActiveSegment.TrailingDiamond, terminal.Background) && !strings.Contains(b.previousActiveSegment.TrailingDiamond, terminal.Foreground) {
		return
	}

	match := regex.FindNamedRegexMatch(terminal.AnchorRegex, b.previousActiveSegment.TrailingDiamond)
	if len(match) == 0 {
		return
	}

	adjustOverride := func(anchor, override string) {
		newOverride := override
		switch override {
		case terminal.Foreground:
			newOverride = terminal.ParentForeground
		case terminal.Background:
			newOverride = terminal.ParentBackground
		}

		if override == newOverride {
			return
		}

		newAnchor := strings.Replace(match[terminal.ANCHOR], override, newOverride, 1)
		b.previousActiveSegment.TrailingDiamond = strings.Replace(b.previousActiveSegment.TrailingDiamond, anchor, newAnchor, 1)
	}

	if len(match[terminal.BG]) > 0 {
		adjustOverride(match[terminal.ANCHOR], match[terminal.BG])
	}
	if len(match[terminal.FG]) > 0 {
		adjustOverride(match[terminal.ANCHOR], match[terminal.FG])
	}
}

func (b *Block) getPowerlineColor() string {
	if b.previousActiveSegment == nil {
		return terminal.Transparent
	}
	if b.previousActiveSegment.style() == Diamond && len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.activeSegment.style() == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if !b.previousActiveSegment.isPowerline() {
		return terminal.Transparent
	}
	return b.previousActiveSegment.background()
}
