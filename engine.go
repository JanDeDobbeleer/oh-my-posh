package main

import (
	"fmt"
	"sync"
)

type engine struct {
	settings              *Settings
	env                   environmentInfo
	renderer              *Renderer
	activeBlock           *Block
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (e *engine) getPowerlineColor(foreground bool) string {
	if e.previousActiveSegment == nil {
		return Transparent
	}
	if !foreground && e.activeSegment.Style != Powerline {
		return Transparent
	}
	if foreground && e.previousActiveSegment.Style != Powerline {
		return Transparent
	}
	return e.previousActiveSegment.Background
}

func (e *engine) writePowerLineSeparator(background string, foreground string, end bool) {
	symbol := e.activeSegment.PowerlineSymbol
	if end {
		symbol = e.previousActiveSegment.PowerlineSymbol
	}
	if e.activeSegment.InvertPowerline {
		e.renderer.write(foreground, background, symbol)
		return
	}
	e.renderer.write(background, foreground, symbol)
}

func (e *engine) endPowerline() {
	if e.activeSegment != nil &&
		e.activeSegment.Style != Powerline &&
		e.previousActiveSegment != nil &&
		e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(e.getPowerlineColor(false), e.previousActiveSegment.Background, true)
	}
}

func (e *engine) renderPowerLineSegment(text string) {
	e.writePowerLineSeparator(e.activeSegment.Background, e.getPowerlineColor(true), false)
	e.renderText(text)
}

func (e *engine) renderPlainSegment(text string) {
	e.renderText(text)
}

func (e *engine) renderDiamondSegment(text string) {
	e.renderer.write(Transparent, e.activeSegment.Background, e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.renderer.write(Transparent, e.activeSegment.Background, e.activeSegment.TrailingDiamond)
}

func (e *engine) renderText(text string) {
	prefix := e.activeSegment.getValue(Prefix, " ")
	postfix := e.activeSegment.getValue(Postfix, " ")
	e.renderer.write(e.activeSegment.Background, e.activeSegment.Foreground, fmt.Sprintf("%s%s%s", prefix, text, postfix))
}

func (e *engine) renderSegmentText(text string) {
	switch e.activeSegment.Style {
	case Plain:
		e.renderPlainSegment(text)
	case Diamond:
		e.renderDiamondSegment(text)
	default:
		e.renderPowerLineSegment(text)
	}
	e.previousActiveSegment = e.activeSegment
}

func (e *engine) renderBlockSegments(block *Block) string {
	defer e.reset()
	e.activeBlock = block
	e.setStringValues(block.Segments)
	for _, segment := range block.Segments {
		if !segment.active {
			continue
		}
		e.activeSegment = segment
		e.endPowerline()
		text := segment.stringValue
		e.activeSegment.Background = segment.props.background
		e.activeSegment.Foreground = segment.props.foreground
		e.renderSegmentText(text)
	}
	if e.previousActiveSegment != nil && e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(Transparent, e.previousActiveSegment.Background, true)
	}
	return e.renderer.string()
}

func (e *engine) setStringValues(segments []*Segment) {
	wg := sync.WaitGroup{}
	wg.Add(len(segments))
	defer wg.Wait()
	cwd := e.env.getcwd()
	for _, segment := range segments {
		go func(s *Segment) {
			defer wg.Done()
			err := s.mapSegmentWithWriter(e.env)
			if err == nil && !s.hasValue(IgnoreFolders, cwd) && s.enabled() {
				s.stringValue = s.string()
			}
		}(segment)
	}
}

func (e *engine) render() {
	for _, block := range e.settings.Blocks {
		// if line break, append a line break
		if block.Type == LineBreak {
			e.renderer.print("\n")
			continue
		}
		if block.VerticalOffset != 0 {
			e.renderer.print(e.renderer.changeLine(block.VerticalOffset))
		}
		switch block.Alignment {
		case Right:
			e.renderer.print(e.renderer.carriageForward())
			blockText := e.renderBlockSegments(block)
			cursorMove := e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
			e.renderer.print(cursorMove)
			e.renderer.print(blockText)
		default:
			e.renderer.print(e.renderBlockSegments(block))
		}
	}
	if e.settings.ConsoleTitle {
		e.renderer.setConsoleTitle(e.env.getcwd())
	}
	e.renderer.creset()
	if e.settings.FinalSpace {
		e.renderer.print(" ")
	}
}

func (e *engine) reset() {
	e.renderer.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
