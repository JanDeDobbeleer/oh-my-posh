package main

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
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

	env                   environment.Environment
	writer                color.Writer
	ansi                  *color.Ansi
	activeSegment         *Segment
	previousActiveSegment *Segment
	activeBackground      string
	activeForeground      string
}

func (b *Block) init(env environment.Environment, writer color.Writer, ansi *color.Ansi) {
	b.env = env
	b.writer = writer
	b.ansi = ansi
}

func (b *Block) initPlain(env environment.Environment, config *Config) {
	b.ansi = &color.Ansi{}
	b.ansi.Init(plain)
	b.writer = &color.AnsiWriter{
		Ansi:               b.ansi,
		TerminalBackground: getConsoleBackgroundColor(env, config.TerminalBackground),
		AnsiColors:         config.MakeColors(env),
	}
	b.env = env
}

func (b *Block) setActiveSegment(segment *Segment) {
	b.activeSegment = segment
	b.activeBackground = segment.background()
	b.activeForeground = segment.foreground()
	b.writer.SetColors(b.activeBackground, b.activeForeground)
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
	defer b.writer.Reset()
	for _, segment := range b.Segments {
		if !segment.active {
			continue
		}
		b.renderSegment(segment)
	}
	b.writePowerline(true)
	b.writer.ClearParentColors()
	return b.writer.String()
}

func (b *Block) renderSegment(segment *Segment) {
	b.setActiveSegment(segment)
	b.writePowerline(false)
	switch b.activeSegment.Style {
	case Plain, Powerline:
		b.renderText(segment.stringValue)
	case Diamond:
		b.writer.Write(color.Transparent, b.activeBackground, b.activeSegment.LeadingDiamond)
		b.renderText(segment.stringValue)
		b.writer.Write(color.Transparent, b.activeBackground, b.activeSegment.TrailingDiamond)
	}
	b.previousActiveSegment = b.activeSegment
	b.writer.SetParentColors(b.activeBackground, b.activeForeground)
}

func (b *Block) renderText(text string) {
	defaultValue := " "
	b.writer.Write(b.activeBackground, b.activeForeground, b.activeSegment.getValue(Prefix, defaultValue))
	b.writer.Write(b.activeBackground, b.activeForeground, text)
	b.writer.Write(b.activeBackground, b.activeForeground, b.activeSegment.getValue(Postfix, defaultValue))
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
		background = color.Transparent
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		background = b.activeSegment.background()
	}
	if b.activeSegment.InvertPowerline {
		b.writer.Write(b.getPowerlineColor(), background, symbol)
		return
	}
	b.writer.Write(background, b.getPowerlineColor(), symbol)
}

func (b *Block) getPowerlineColor() string {
	if b.previousActiveSegment == nil {
		return color.Transparent
	}
	if b.previousActiveSegment.Style == Diamond && len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.previousActiveSegment.Style != Powerline {
		return color.Transparent
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
			segmentTiming.stringValue = b.writer.String()
			b.writer.Reset()
		}
		segmentTimings = append(segmentTimings, &segmentTiming)
	}
	return largestSegmentNameLength, segmentTimings
}
