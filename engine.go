package main

import (
	"fmt"
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
	cwd, _ := e.env.getwd()
	for _, segment := range block.Segments {
		if segment.hasValue(IgnoreFolders, cwd) {
			continue
		}
		props, err := segment.mapSegmentWithWriter(e.env)
		if err != nil || !segment.enabled() {
			continue
		}
		e.activeSegment = segment
		e.endPowerline()
		text := segment.string()
		e.activeSegment.Background = props.background
		e.activeSegment.Foreground = props.foreground
		e.renderSegmentText(text)
	}
	if e.previousActiveSegment != nil && e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(Transparent, e.previousActiveSegment.Background, true)
	}
	return e.renderer.string()
}

func (e *engine) render() {
	for _, block := range e.settings.Blocks {
		// if line break, append a line break
		if block.Type == LineBreak {
			fmt.Print(e.renderer.lineBreak())
			continue
		}
		if block.VerticalOffset != 0 {
			fmt.Print(e.renderer.changeLine(block.VerticalOffset))
		}
		switch block.Alignment {
		case Right:
			fmt.Print(e.renderer.carriageForward())
			blockText := e.renderBlockSegments(block)
			cursorMove := e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
			fmt.Print(cursorMove)
			fmt.Print(blockText)
		default:
			fmt.Print(e.renderBlockSegments(block))
		}
	}
	if e.settings.FinalSpace {
		fmt.Print(" ")
	}
}

func (e *engine) reset() {
	e.renderer.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
