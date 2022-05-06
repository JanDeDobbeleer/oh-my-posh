package engine

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
	"oh-my-posh/shell"
	"strings"
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
	Type             BlockType      `json:"type,omitempty"`
	Alignment        BlockAlignment `json:"alignment,omitempty"`
	HorizontalOffset int            `json:"horizontal_offset,omitempty"`
	VerticalOffset   int            `json:"vertical_offset,omitempty"`
	Segments         []*Segment     `json:"segments,omitempty"`
	Newline          bool           `json:"newline,omitempty"`
	Filler           string         `json:"filler,omitempty"`

	env                   environment.Environment
	writer                color.Writer
	ansi                  *color.Ansi
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (b *Block) Init(env environment.Environment, writer color.Writer, ansi *color.Ansi) {
	b.env = env
	b.writer = writer
	b.ansi = ansi
	b.setEnabledSegments()
	b.setSegmentsText()
}

func (b *Block) InitPlain(env environment.Environment, config *Config) {
	b.ansi = &color.Ansi{}
	b.ansi.InitPlain(env.Shell())
	b.writer = &color.AnsiWriter{
		Ansi:               b.ansi,
		TerminalBackground: shell.ConsoleBackgroundColor(env, config.TerminalBackground),
		AnsiColors:         config.MakeColors(env),
	}
	b.env = env
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
		if segment.enabled {
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
			s.setEnabled(b.env)
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
			if !s.enabled {
				return
			}
			s.text = s.string()
			s.enabled = len(strings.ReplaceAll(s.text, " ", "")) > 0
		}(segment)
	}
}

func (b *Block) renderSegments() (string, int) {
	defer b.writer.Reset()
	for _, segment := range b.Segments {
		if !segment.enabled && segment.Style != Accordion {
			continue
		}
		b.setActiveSegment(segment)
		b.renderActiveSegment()
	}
	b.writePowerline(true)
	b.writer.ClearParentColors()
	return b.writer.String()
}

func (b *Block) renderActiveSegment() {
	b.writePowerline(false)
	switch b.activeSegment.Style {
	case Plain, Powerline:
		b.writer.Write(color.Background, color.Foreground, b.activeSegment.text)
	case Diamond:
		b.writer.Write(color.Transparent, color.Background, b.activeSegment.LeadingDiamond)
		b.writer.Write(color.Background, color.Foreground, b.activeSegment.text)
		b.writer.Write(color.Transparent, color.Background, b.activeSegment.TrailingDiamond)
	case Accordion:
		if b.activeSegment.enabled {
			b.writer.Write(color.Background, color.Foreground, b.activeSegment.text)
		}
	}
	b.previousActiveSegment = b.activeSegment
	b.writer.SetParentColors(b.previousActiveSegment.background(), b.previousActiveSegment.foreground())
}

func (b *Block) writePowerline(final bool) {
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
	bgColor := color.Background
	if final || !b.activeSegment.isPowerline() {
		bgColor = color.Transparent
	}
	if b.activeSegment.Style == Diamond && len(b.activeSegment.LeadingDiamond) == 0 {
		bgColor = color.Background
	}
	if b.activeSegment.InvertPowerline {
		b.writer.Write(b.getPowerlineColor(), bgColor, symbol)
		return
	}
	b.writer.Write(bgColor, b.getPowerlineColor(), symbol)
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
	if !b.previousActiveSegment.isPowerline() {
		return color.Transparent
	}
	return b.previousActiveSegment.background()
}

func (b *Block) debug() (int, []*SegmentTiming) {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	for _, segment := range b.Segments {
		var segmentTiming SegmentTiming
		segmentTiming.name = string(segment.Type)
		segmentTiming.nameLength = len(segmentTiming.name)
		if segmentTiming.nameLength > largestSegmentNameLength {
			largestSegmentNameLength = segmentTiming.nameLength
		}
		start := time.Now()
		segmentTiming.active = segment.enabled
		if segmentTiming.active || segment.Style == Accordion {
			b.setActiveSegment(segment)
			b.renderActiveSegment()
			segmentTiming.text, _ = b.writer.String()
			b.writer.Reset()
		}
		segmentTiming.duration = time.Since(start)
		segmentTimings = append(segmentTimings, &segmentTiming)
	}
	return largestSegmentNameLength, segmentTimings
}
