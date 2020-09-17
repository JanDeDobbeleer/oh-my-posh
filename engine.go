package main

import (
	"fmt"
	"regexp"

	"golang.org/x/text/unicode/norm"
)

type engine struct {
	settings              *Settings
	env                   environmentInfo
	renderer              *ColorWriter
	activeBlock           *Block
	activeSegment         *Segment
	previousActiveSegment *Segment
}

func (e *engine) getPowerlineColor(foreground bool) string {
	if e.previousActiveSegment == nil {
		return e.settings.ConsoleBackgroundColor
	}
	if !foreground && e.activeSegment.Style != Powerline {
		return e.settings.ConsoleBackgroundColor
	}
	if foreground && e.previousActiveSegment.Style != Powerline {
		return e.settings.ConsoleBackgroundColor
	}
	return e.previousActiveSegment.Background
}

func (e *engine) writePowerLineSeparator(background string, foreground string) {
	if e.activeBlock.InvertPowerlineSeparatorColor {
		e.renderer.write(foreground, background, e.activeBlock.PowerlineSeparator)
		return
	}
	e.renderer.write(background, foreground, e.activeBlock.PowerlineSeparator)
}

func (e *engine) endPowerline() {
	if e.activeSegment != nil &&
		e.activeSegment.Style != Powerline &&
		e.previousActiveSegment != nil &&
		e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(e.getPowerlineColor(false), e.previousActiveSegment.Background)
	}
}

func (e *engine) renderPowerLineSegment(text string) {
	e.writePowerLineSeparator(e.activeSegment.Background, e.getPowerlineColor(true))
	e.renderText(text)
}

func (e *engine) renderPlainSegment(text string) {
	e.renderText(text)
}

func (e *engine) renderDiamondSegment(text string) {
	e.renderer.write(e.settings.ConsoleBackgroundColor, e.activeSegment.Background, e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.renderer.write(e.settings.ConsoleBackgroundColor, e.activeSegment.Background, e.activeSegment.TrailingDiamond)
}

func (e *engine) getStringProperty(property Property, defaultValue string) string {
	if value, ok := e.activeSegment.Properties[property]; ok {
		return parseString(value, defaultValue)
	}
	return defaultValue
}

func (e *engine) renderText(text string) {
	prefix := e.getStringProperty(Prefix, " ")
	postfix := e.getStringProperty(Postfix, " ")
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
	for _, segment := range block.Segments {
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
		e.writePowerLineSeparator(e.settings.ConsoleBackgroundColor, e.previousActiveSegment.Background)
	}
	return e.renderer.string()
}

func (e *engine) lenWithoutANSI(str string) int {
	ansi := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	re := regexp.MustCompile(ansi)
	stripped := re.ReplaceAllString(str, "")
	var i norm.Iter
	i.InitString(norm.NFD, stripped)
	var count int
	for !i.Done() {
		i.Next()
		count++
	}
	return count
}

func (e *engine) render() {
	for _, block := range e.settings.Blocks {
		// if line break, append a line break
		if block.Type == LineBreak {
			fmt.Printf("\x1b[%dC ", 1000)
			continue
		}
		if block.LineOffset < 0 {
			fmt.Printf("\x1b[%dF", -block.LineOffset)
		} else if block.LineOffset > 0 {
			fmt.Printf("\x1b[%dB", block.LineOffset)
		}
		switch block.Alignment {
		case Right:
			fmt.Printf("\x1b[%dC", 1000)
			blockText := e.renderBlockSegments(block)
			fmt.Printf("\x1b[%dD", e.lenWithoutANSI(blockText)+e.settings.RightSegmentOffset)
			fmt.Print(blockText)
		default:
			fmt.Print(e.renderBlockSegments(block))
		}
	}
	if e.settings.EndSpaceEnabled {
		fmt.Print(" ")
	}
}

func (e *engine) reset() {
	e.renderer.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
