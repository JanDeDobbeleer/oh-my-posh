package main

import (
	"fmt"
	"sync"
)

type engine struct {
	settings              *Settings
	env                   environmentInfo
	color                 *AnsiColor
	renderer              *AnsiRenderer
	activeBlock           *Block
	activeSegment         *Segment
	previousActiveSegment *Segment
	rprompt               string
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

func (e *engine) writePowerLineSeparator(background, foreground string, end bool) {
	symbol := e.activeSegment.PowerlineSymbol
	if end {
		symbol = e.previousActiveSegment.PowerlineSymbol
	}
	if e.activeSegment.InvertPowerline {
		e.color.write(foreground, background, symbol)
		return
	}
	e.color.write(background, foreground, symbol)
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
	e.color.write(Transparent, e.activeSegment.Background, e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.color.write(Transparent, e.activeSegment.Background, e.activeSegment.TrailingDiamond)
}

func (e *engine) renderText(text string) {
	defaultValue := " "
	if e.activeSegment.Background != "" {
		defaultValue = fmt.Sprintf("<%s>\u2588</>", e.activeSegment.Background)
	}
	prefix := e.activeSegment.getValue(Prefix, defaultValue)
	postfix := e.activeSegment.getValue(Postfix, defaultValue)
	e.color.write(e.activeSegment.Background, e.activeSegment.Foreground, fmt.Sprintf("%s%s%s", prefix, text, postfix))
	if *e.env.getArgs().Debug {
		e.color.write(e.activeSegment.Background, e.activeSegment.Foreground, fmt.Sprintf("(%s:%s)", e.activeSegment.Type, e.activeSegment.timing))
	}
}

func (e *engine) renderSegmentText(text string) {
	switch e.activeSegment.Style {
	case Plain:
		e.renderPlainSegment(text)
	case Diamond:
		e.renderDiamondSegment(text)
	case Powerline:
		e.renderPowerLineSegment(text)
	}
	e.previousActiveSegment = e.activeSegment
}

func (e *engine) renderBlockSegments(block *Block) string {
	defer e.resetBlock()
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
	return e.color.string()
}

func (e *engine) setStringValues(segments []*Segment) {
	wg := sync.WaitGroup{}
	wg.Add(len(segments))
	defer wg.Wait()
	cwd := e.env.getcwd()
	debug := *e.env.getArgs().Debug
	for _, segment := range segments {
		go func(s *Segment) {
			defer wg.Done()
			s.setStringValue(e.env, cwd, debug)
		}(segment)
	}
}

func (e *engine) render() {
	for _, block := range e.settings.Blocks {
		// if line break, append a line break
		switch block.Type {
		case LineBreak:
			e.renderer.print("\n")
		case Prompt:
			if block.VerticalOffset != 0 {
				e.renderer.changeLine(block.VerticalOffset)
			}
			switch block.Alignment {
			case Right:
				e.renderer.carriageForward()
				blockText := e.renderBlockSegments(block)
				e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
				e.renderer.print(blockText)
			case Left:
				e.renderer.print(e.renderBlockSegments(block))
			}
		case RPrompt:
			e.rprompt = e.renderBlockSegments(block)
		}
	}
	if e.settings.ConsoleTitle {
		switch e.settings.ConsoleTitleStyle {
		case FullPath:
			e.renderer.setConsoleTitle(e.env.getcwd())
		case FolderName:
			fallthrough
		default:
			e.renderer.setConsoleTitle(base(e.env.getcwd(), e.env))
		}
	}
	e.renderer.creset()
	if e.settings.FinalSpace {
		e.renderer.print(" ")
	}
	e.write()
}

func (e *engine) write() {
	if *e.env.getArgs().Eval {
		fmt.Printf("PS1=\"%s\"", e.renderer.string())
		if e.env.getShellName() == zsh {
			fmt.Printf("\nRPROMPT=\"%s\"", e.rprompt)
		}
		return
	}

	if e.rprompt != "" && (e.env.getShellName() == pwsh || e.env.getShellName() == powershell5) {
		e.renderer.saveCursorPosition()
		e.renderer.carriageForward()
		e.renderer.setCursorForRightWrite(e.rprompt, 0)
		e.renderer.print(e.rprompt)
		e.renderer.restoreCursorPosition()
	}
	fmt.Print(e.renderer.string())
}

func (e *engine) resetBlock() {
	e.color.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
