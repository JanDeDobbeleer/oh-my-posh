package main

import (
	"sync"
	"time"
)

// BlockType type of block
type BlockType string

// BlockAlignment aligment of a Block
type BlockAlignment string

const (
	// Prompt writes one or more Segments
	Prompt BlockType = "prompt"
	// LineBreak creates a line break in the prompt
	LineBreak BlockType = "newline"
	// RPrompt a right aligned prompt in ZSH and Powershell
	RPrompt BlockType = "rprompt"
	// Left aligns left
	Left BlockAlignment = "left"
	// Right aligns right
	Right BlockAlignment = "right"
)

// Block defines a part of the prompt with optional segments
type Block struct {
	Type             BlockType      `config:"type"`
	Alignment        BlockAlignment `config:"alignment"`
	HorizontalOffset int            `config:"horizontal_offset"`
	VerticalOffset   int            `config:"vertical_offset"`
	Segments         []*Segment     `config:"segments"`
	Newline          bool           `config:"newline"`

	env                   Environment
	writer                promptWriter
	ansi                  *ansiUtils
	activeSegment         *Segment
	previousActiveSegment *Segment
	activeBackground      string
	activeForeground      string
}

func (b *Block) init(env Environment, writer promptWriter, ansi *ansiUtils) {
	b.env = env
	b.writer = writer
	b.ansi = ansi
}

func (b *Block) initPlain(env Environment, config *Config) {
	b.ansi = &ansiUtils{}
	b.ansi.init(plain)
	b.writer = &AnsiWriter{
		ansi:               b.ansi,
		terminalBackground: getConsoleBackgroundColor(env, config.TerminalBackground),
		ansiColors:         MakeColors(env, config),
	}
	b.env = env
}

func (b *Block) setActiveSegment(segment *Segment) {
	b.activeSegment = segment
	b.activeBackground = segment.background()
	b.activeForeground = segment.foreground()
	b.writer.setColors(b.activeBackground, b.activeForeground)
}

func (b *Block) enabled() bool {
	if b.Type == LineBreak {
		return true
	}
	for _, segment := range b.Segments {
		if segment.active {
			return true
		}
	}
	return false
}

func (b *Block) setStringValues() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.Segments))
	defer wg.Wait()
	for _, segment := range b.Segments {
		go func(s *Segment) {
			defer wg.Done()
			s.setStringValue(b.env)
		}(segment)
	}
}

func (b *Block) renderSegments() string {
	defer b.writer.reset()
	for _, segment := range b.Segments {
		if !segment.active {
			continue
		}
		b.renderSegment(segment)
	}
	b.writePowerline(true)
	b.writer.clearParentColors()
	return b.writer.string()
}

func (b *Block) renderSegment(segment *Segment) {
	b.setActiveSegment(segment)
	b.writePowerline(false)
	switch b.activeSegment.Style {
	case Plain, Powerline:
		b.renderText(segment.stringValue)
	case Diamond:
		b.writer.write(Transparent, b.activeBackground, b.activeSegment.LeadingDiamond)
		b.renderText(segment.stringValue)
		b.writer.write(Transparent, b.activeBackground, b.activeSegment.TrailingDiamond)
	}
	b.previousActiveSegment = b.activeSegment
	b.writer.setParentColors(b.activeBackground, b.activeForeground)
}

func (b *Block) renderText(text string) {
	defaultValue := " "
	b.writer.write(b.activeBackground, b.activeForeground, b.activeSegment.getValue(Prefix, defaultValue))
	b.writer.write(b.activeBackground, b.activeForeground, text)
	b.writer.write(b.activeBackground, b.activeForeground, b.activeSegment.getValue(Postfix, defaultValue))
}

func (b *Block) writePowerline(final bool) {
	resolvePowerlineSymbol := func() string {
		var symbol string
		if b.activeSegment.Style == Powerline {
			symbol = b.activeSegment.PowerlineSymbol
		} else if b.previousActiveSegment != nil && b.previousActiveSegment.Style == Powerline {
			symbol = b.previousActiveSegment.PowerlineSymbol
		}
		return symbol
	}
	symbol := resolvePowerlineSymbol()
	if len(symbol) == 0 {
		return
	}
	background := b.activeSegment.background()
	if final || b.activeSegment.Style != Powerline {
		background = Transparent
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		background = b.activeSegment.background()
	}
	if b.activeSegment.InvertPowerline {
		b.writer.write(b.getPowerlineColor(), background, symbol)
		return
	}
	b.writer.write(background, b.getPowerlineColor(), symbol)
}

func (b *Block) getPowerlineColor() string {
	if b.previousActiveSegment == nil {
		return Transparent
	}
	if b.previousActiveSegment.Style == Diamond && len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.previousActiveSegment.Style != Powerline {
		return Transparent
	}
	return b.previousActiveSegment.background()
}

func (b *Block) debug() (int, []*SegmentTiming) {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	for _, segment := range b.Segments {
		err := segment.mapSegmentWithWriter(b.env)
		if err != nil || !segment.shouldIncludeFolder() {
			continue
		}
		var segmentTiming SegmentTiming
		segmentTiming.name = string(segment.Type)
		segmentTiming.nameLength = len(segmentTiming.name)
		if segmentTiming.nameLength > largestSegmentNameLength {
			largestSegmentNameLength = segmentTiming.nameLength
		}
		// enabled() timing
		start := time.Now()
		segmentTiming.enabled = segment.enabled()
		segmentTiming.enabledDuration = time.Since(start)
		// string() timing
		if segmentTiming.enabled {
			start = time.Now()
			segment.stringValue = segment.string()
			segmentTiming.stringDuration = time.Since(start)
			b.renderSegment(segment)
			b.writePowerline(true)
			segmentTiming.stringValue = b.writer.string()
			b.writer.reset()
		}
		segmentTimings = append(segmentTimings, &segmentTiming)
	}
	return largestSegmentNameLength, segmentTimings
}
