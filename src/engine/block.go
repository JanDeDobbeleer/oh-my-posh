package engine

import (
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
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
	Type      BlockType      `json:"type,omitempty"`
	Alignment BlockAlignment `json:"alignment,omitempty"`
	Segments  []*Segment     `json:"segments,omitempty"`
	Newline   bool           `json:"newline,omitempty"`
	Filler    string         `json:"filler,omitempty"`
	Overflow  Overflow       `json:"overflow,omitempty"`

	// Deprecated: keep the logic for legacy purposes
	HorizontalOffset int `json:"horizontal_offset,omitempty"`
	VerticalOffset   int `json:"vertical_offset,omitempty"`

	MaxWidth int `json:"max_width,omitempty"`
	MinWidth int `json:"min_width,omitempty"`

	env                   platform.Environment
	writer                *ansi.Writer
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (b *Block) Init(env platform.Environment, writer *ansi.Writer) {
	b.env = env
	b.writer = writer
	b.executeSegmentLogic()
}

func (b *Block) InitPlain(env platform.Environment, config *Config) {
	b.writer = &ansi.Writer{
		TerminalBackground: shell.ConsoleBackgroundColor(env, config.TerminalBackground),
		AnsiColors:         config.MakeColors(),
		TrueColor:          env.Flags().TrueColor,
	}
	b.writer.Init(shell.GENERIC)
	b.env = env
	b.executeSegmentLogic()
}

func (b *Block) executeSegmentLogic() {
	if b.env.Flags().Debug {
		return
	}
	if shouldHideForWidth(b.env, b.MinWidth, b.MaxWidth) {
		return
	}
	b.setEnabledSegments()
	b.setSegmentsText()
}

func (b *Block) setActiveSegment(segment *Segment) {
	b.activeSegment = segment
	b.writer.SetColors(segment.background(), segment.foreground())
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
	for _, segment := range b.Segments {
		if !segment.Enabled && segment.style() != Accordion {
			continue
		}
		if colors, newCycle := cycle.Loop(); colors != nil {
			cycle = &newCycle
			segment.colors = colors
		}
		b.setActiveSegment(segment)
		b.renderActiveSegment()
	}
	b.writeSeparator(true)
	return b.writer.String()
}

func (b *Block) renderActiveSegment() {
	b.writeSeparator(false)
	switch b.activeSegment.style() {
	case Plain, Powerline:
		b.writer.Write(ansi.Background, ansi.Foreground, b.activeSegment.text)
	case Diamond:
		b.writer.Write(ansi.Transparent, ansi.Background, b.activeSegment.LeadingDiamond)
		b.writer.Write(ansi.Background, ansi.Foreground, b.activeSegment.text)
	case Accordion:
		if b.activeSegment.Enabled {
			b.writer.Write(ansi.Background, ansi.Foreground, b.activeSegment.text)
		}
	}
	b.previousActiveSegment = b.activeSegment
	b.writer.SetParentColors(b.previousActiveSegment.background(), b.previousActiveSegment.foreground())
}

func (b *Block) writeSeparator(final bool) {
	isCurrentDiamond := b.activeSegment.style() == Diamond
	if final && isCurrentDiamond {
		b.writer.Write(ansi.Transparent, ansi.Background, b.activeSegment.TrailingDiamond)
		return
	}
	isPreviousDiamond := b.previousActiveSegment != nil && b.previousActiveSegment.style() == Diamond
	if isPreviousDiamond && isCurrentDiamond && len(b.activeSegment.LeadingDiamond) == 0 {
		b.writer.Write(ansi.Background, ansi.ParentBackground, b.previousActiveSegment.TrailingDiamond)
		return
	}
	if isPreviousDiamond && len(b.previousActiveSegment.TrailingDiamond) > 0 {
		b.writer.Write(ansi.Transparent, ansi.ParentBackground, b.previousActiveSegment.TrailingDiamond)
	}

	resolvePowerlineSymbol := func() string {
		var symbol string
		if b.activeSegment.isPowerline() {
			symbol = b.activeSegment.PowerlineSymbol
		} else if b.previousActiveSegment != nil && b.previousActiveSegment.isPowerline() {
			symbol = b.previousActiveSegment.PowerlineSymbol
		}
		return symbol
	}
	symbol := resolvePowerlineSymbol()
	if len(symbol) == 0 {
		return
	}
	bgColor := ansi.Background
	if final || !b.activeSegment.isPowerline() {
		bgColor = ansi.Transparent
	}
	if b.activeSegment.style() == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		bgColor = ansi.Background
	}
	if b.activeSegment.InvertPowerline {
		b.writer.Write(b.getPowerlineColor(), bgColor, symbol)
		return
	}
	b.writer.Write(bgColor, b.getPowerlineColor(), symbol)
}

func (b *Block) getPowerlineColor() string {
	if b.previousActiveSegment == nil {
		return ansi.Transparent
	}
	if b.previousActiveSegment.style() == Diamond && len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.activeSegment.style() == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if !b.previousActiveSegment.isPowerline() {
		return ansi.Transparent
	}
	return b.previousActiveSegment.background()
}

func (b *Block) Debug() (int, []*SegmentTiming) {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	for _, segment := range b.Segments {
		var segmentTiming SegmentTiming
		segmentTiming.name = string(segment.Type)
		segmentTiming.nameLength = len(segmentTiming.name)
		if segmentTiming.nameLength > largestSegmentNameLength {
			largestSegmentNameLength = segmentTiming.nameLength
		}
		b.env.DebugF("Segment: %s", segmentTiming.name)
		start := time.Now()
		segment.SetEnabled(b.env)
		segment.SetText()
		segmentTiming.active = segment.Enabled
		if segmentTiming.active || segment.style() == Accordion {
			b.setActiveSegment(segment)
			b.renderActiveSegment()
			segmentTiming.text, _ = b.writer.String()
		}
		segmentTiming.duration = time.Since(start)
		segmentTimings = append(segmentTimings, &segmentTiming)
	}
	return largestSegmentNameLength, segmentTimings
}
