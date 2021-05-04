package main

import (
	"fmt"
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

	env                   environmentInfo
	writer                colorWriter
	ansi                  *ansiUtils
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (b *Block) init(env environmentInfo, writer colorWriter, ansi *ansiUtils) {
	b.env = env
	b.writer = writer
	b.ansi = ansi
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
	cwd := b.env.getcwd()
	for _, segment := range b.Segments {
		go func(s *Segment) {
			defer wg.Done()
			s.setStringValue(b.env, cwd)
		}(segment)
	}
}

func (b *Block) renderSegments() string {
	defer b.writer.reset()
	for _, segment := range b.Segments {
		if !segment.active {
			continue
		}
		b.activeSegment = segment
		b.endPowerline()
		b.renderSegmentText(segment.stringValue)
	}
	if b.previousActiveSegment != nil && b.previousActiveSegment.Style == Powerline {
		b.writePowerLineSeparator(Transparent, b.previousActiveSegment.background(), true)
	}
	return b.writer.string()
}

func (b *Block) endPowerline() {
	if b.previousActiveSegment == nil || b.activeSegment == nil {
		return
	}
	if b.activeSegment.Style != Powerline &&
		b.previousActiveSegment.Style == Powerline {
		b.writePowerLineSeparator(b.getPowerlineColor(false), b.previousActiveSegment.background(), true)
	}
}

func (b *Block) writePowerLineSeparator(background, foreground string, end bool) {
	symbol := b.activeSegment.PowerlineSymbol
	if end {
		symbol = b.previousActiveSegment.PowerlineSymbol
	}
	if b.activeSegment.InvertPowerline {
		b.writer.write(foreground, background, symbol)
		return
	}
	b.writer.write(background, foreground, symbol)
}

func (b *Block) getPowerlineColor(foreground bool) string {
	if b.previousActiveSegment == nil {
		return Transparent
	}
	if b.previousActiveSegment.Style == Diamond && len(b.previousActiveSegment.TrailingDiamond) == 0 {
		return b.previousActiveSegment.background()
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		return b.activeSegment.background()
	}
	if !foreground && b.activeSegment.Style != Powerline {
		return Transparent
	}
	if foreground && b.previousActiveSegment.Style != Powerline {
		return Transparent
	}
	return b.previousActiveSegment.background()
}

func (b *Block) renderSegmentText(text string) {
	switch b.activeSegment.Style {
	case Plain:
		b.renderPlainSegment(text)
	case Diamond:
		b.renderDiamondSegment(text)
	case Powerline:
		b.renderPowerLineSegment(text)
	}
	b.previousActiveSegment = b.activeSegment
}

func (b *Block) renderPowerLineSegment(text string) {
	b.writePowerLineSeparator(b.activeSegment.background(), b.getPowerlineColor(true), false)
	b.renderText(text)
}

func (b *Block) renderPlainSegment(text string) {
	b.renderText(text)
}

func (b *Block) renderDiamondSegment(text string) {
	b.writer.write(Transparent, b.activeSegment.background(), b.activeSegment.LeadingDiamond)
	b.renderText(text)
	b.writer.write(Transparent, b.activeSegment.background(), b.activeSegment.TrailingDiamond)
}

func (b *Block) renderText(text string) {
	text = b.ansi.generateHyperlink(text)
	defaultValue := " "
	prefix := b.activeSegment.getValue(Prefix, defaultValue)
	postfix := b.activeSegment.getValue(Postfix, defaultValue)
	b.writer.write(b.activeSegment.background(), b.activeSegment.foreground(), fmt.Sprintf("%s%s%s", prefix, text, postfix))
}

func (b *Block) debug() (int, []*SegmentTiming) {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	for _, segment := range b.Segments {
		err := segment.mapSegmentWithWriter(b.env)
		if err != nil || !segment.shouldIncludeFolder(b.env.getcwd()) {
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
			segmentTiming.stringValue = segment.string()
			segmentTiming.stringDuration = time.Since(start)
			b.previousActiveSegment = nil
			b.activeSegment = segment
			b.renderSegmentText(segmentTiming.stringValue)
			if b.activeSegment.Style == Powerline {
				b.writePowerLineSeparator(Transparent, b.activeSegment.background(), true)
			}
			segmentTiming.stringValue = b.writer.string()
			b.writer.reset()
		}
		segmentTimings = append(segmentTimings, &segmentTiming)
	}
	return largestSegmentNameLength, segmentTimings
}
