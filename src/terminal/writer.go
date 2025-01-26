package terminal

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/mattn/go-runewidth"
)

func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}

type style struct {
	AnchorStart string
	AnchorEnd   string
	Start       string
	End         string
}

var (
	knownStyles = []*style{
		{AnchorStart: `<b>`, AnchorEnd: `</b>`, Start: "\x1b[1m", End: "\x1b[22m"},
		{AnchorStart: `<u>`, AnchorEnd: `</u>`, Start: "\x1b[4m", End: "\x1b[24m"},
		{AnchorStart: `<o>`, AnchorEnd: `</o>`, Start: "\x1b[53m", End: "\x1b[55m"},
		{AnchorStart: `<i>`, AnchorEnd: `</i>`, Start: "\x1b[3m", End: "\x1b[23m"},
		{AnchorStart: `<s>`, AnchorEnd: `</s>`, Start: "\x1b[9m", End: "\x1b[29m"},
		{AnchorStart: `<d>`, AnchorEnd: `</d>`, Start: "\x1b[2m", End: "\x1b[22m"},
		{AnchorStart: `<f>`, AnchorEnd: `</f>`, Start: "\x1b[5m", End: "\x1b[25m"},
		{AnchorStart: `<r>`, AnchorEnd: `</r>`, Start: "\x1b[7m", End: "\x1b[27m"},
	}

	resetStyle      = &style{AnchorStart: "RESET", AnchorEnd: `</>`, End: "\x1b[0m"}
	backgroundStyle = &style{AnchorStart: "BACKGROUND", AnchorEnd: `</>`, End: "\x1b[49m"}

	BackgroundColor color.Ansi
	CurrentColors   *color.Set
	ParentColors    []*color.Set
	Colors          color.String

	Plain       bool
	Interactive bool

	builder strings.Builder
	length  int

	foregroundColor color.Ansi
	backgroundColor color.Ansi
	currentColor    color.History
	runes           []rune

	isTransparent bool
	isInvisible   bool
	isHyperlink   bool

	Shell   string
	Program string

	formats *shell.Formats
)

const (
	AnchorRegex      = `^(?P<ANCHOR><(?P<FG>[^,<>]+)?,?(?P<BG>[^<>]+)?>)`
	colorise         = "\x1b[%sm"
	transparentStart = "\x1b[0m\x1b[%s;49m\x1b[7m"
	transparentEnd   = "\x1b[27m"
	backgroundEnd    = "\x1b[49m"

	AnsiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	OSC99 = "osc99"
	OSC7  = "osc7"
	OSC51 = "osc51"

	ANCHOR = "ANCHOR"
	BG     = "BG"
	FG     = "FG"

	hyperLinkStart   = "<LINK>"
	hyperLinkEnd     = "</LINK>"
	hyperLinkText    = "<TEXT>"
	hyperLinkTextEnd = "</TEXT>"

	empty = "<>"

	startProgress = "\x1b]9;4;3;0\x07"
	endProgress   = "\x1b]9;4;0;0\x07"

	WindowsTerminal = "Windows Terminal"
	Warp            = "WarpTerminal"
	ITerm           = "iTerm.app"
	AppleTerminal   = "Apple_Terminal"
	Unknown         = "Unknown"
)

func Init(sh string) {
	Shell = sh
	Program = getTerminalName()

	log.Debug("terminal program:", Program)
	log.Debug("terminal shell:", Shell)

	color.TrueColor = Program != AppleTerminal

	formats = shell.GetFormats(Shell)
}

func getTerminalName() string {
	Program = os.Getenv("TERM_PROGRAM")
	if len(Program) != 0 {
		return Program
	}

	wtSession := os.Getenv("WT_SESSION")
	if len(wtSession) != 0 {
		return WindowsTerminal
	}

	return Unknown
}

func SetColors(background, foreground color.Ansi) {
	CurrentColors = &color.Set{
		Background: background,
		Foreground: foreground,
	}
}

func SetParentColors(background, foreground color.Ansi) {
	if ParentColors == nil {
		ParentColors = make([]*color.Set, 0)
	}

	ParentColors = append([]*color.Set{{
		Background: background,
		Foreground: foreground,
	}}, ParentColors...)
}

func ChangeLine(numberOfLines int) string {
	if Plain {
		return ""
	}

	position := "B"

	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}

	return fmt.Sprintf(formats.Linechange, numberOfLines, position)
}

func Pwd(pwdType, userName, hostName, pwd string) string {
	if Plain {
		return ""
	}

	switch pwdType {
	case OSC7:
		return fmt.Sprintf(formats.Osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(formats.Osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(formats.Osc99, pwd)
	}
}

func ClearAfter() string {
	if Plain {
		return ""
	}

	return formats.ClearLine + formats.ClearBelow
}

func FormatTitle(title string) string {
	switch Shell {
	// These shells don't support setting the console title.
	case shell.ELVISH, shell.XONSH, shell.TCSH:
		return ""
	case shell.BASH, shell.ZSH:
		title = trimAnsi(title)
		s := new(strings.Builder)

		// We have to do this to prevent the shell from misidentifying escape sequences.
		for _, char := range title {
			escaped, shouldEscape := formats.EscapeSequences[char]
			if shouldEscape {
				s.WriteString(escaped)
				continue
			}

			s.WriteRune(char)
		}

		return fmt.Sprintf(formats.Title, s.String())
	default:
		return fmt.Sprintf(formats.Title, trimAnsi(title))
	}
}

func EscapeText(text string) string {
	return fmt.Sprintf(formats.Escape, text)
}

func SaveCursorPosition() string {
	return formats.SaveCursorPosition
}

func RestoreCursorPosition() string {
	return formats.RestoreCursorPosition
}

func PromptStart() string {
	return fmt.Sprintf(formats.Escape, "\x1b]133;A\007")
}

func CommandStart() string {
	return fmt.Sprintf(formats.Escape, "\x1b]133;B\007")
}

func CommandFinished(code int, ignore bool) string {
	if ignore {
		return fmt.Sprintf(formats.Escape, "\x1b]133;D\007")
	}

	mark := fmt.Sprintf("\x1b]133;D;%d\007", code)

	return fmt.Sprintf(formats.Escape, mark)
}

func LineBreak() string {
	cr := fmt.Sprintf(formats.Left, 1000)
	lf := fmt.Sprintf(formats.Linechange, 1, "B")
	return cr + lf
}

func StartProgress() string {
	if Program != WindowsTerminal {
		return ""
	}

	return startProgress
}

func StopProgress() string {
	if Program != WindowsTerminal {
		return ""
	}

	return endProgress
}

func Write(background, foreground color.Ansi, text string) {
	if len(text) == 0 {
		return
	}

	backgroundColor, foregroundColor = asAnsiColors(background, foreground)

	// default to white foreground
	if foregroundColor.IsEmpty() {
		foregroundColor = Colors.ToAnsi("white", false)
	}

	// validate if we start with a color override
	match := regex.FindNamedRegexMatch(AnchorRegex, text)
	if len(match) != 0 && match[ANCHOR] != hyperLinkStart {
		colorOverride := true
		for _, style := range knownStyles {
			if match[ANCHOR] != style.AnchorStart {
				continue
			}

			writeEscapedAnsiString(style.Start)
			colorOverride = false
		}

		if colorOverride {
			currentColor.Add(asAnsiColors(color.Ansi(match[BG]), color.Ansi(match[FG])))
		}
	}

	writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match[ANCHOR] == hyperLinkStart {
		isHyperlink = true
		builder.WriteString(formats.HyperlinkStart)
	}

	text = text[len(match[ANCHOR]):]
	runes = []rune(text)
	hyperlinkTextPosition := 0

	for i := 0; i < len(runes); i++ {
		s := runes[i]
		// ignore everything which isn't overriding
		if s != '<' {
			write(s)
			continue
		}

		// color/end overrides first
		text = string(runes[i:])
		match = regex.FindNamedRegexMatch(AnchorRegex, text)
		if len(match) > 0 {
			// check for hyperlinks first
			switch match[ANCHOR] {
			case hyperLinkStart:
				isHyperlink = true
				i += len([]rune(match[ANCHOR])) - 1
				builder.WriteString(formats.HyperlinkStart)
				continue
			case hyperLinkText:
				isHyperlink = false
				i += len([]rune(match[ANCHOR])) - 1
				hyperlinkTextPosition = i
				builder.WriteString(formats.HyperlinkCenter)
				continue
			case hyperLinkTextEnd:
				// this implies there's no text in the hyperlink
				if hyperlinkTextPosition+1 == i {
					builder.WriteString("link")
					length += 4
				}
				i += len([]rune(match[ANCHOR])) - 1
				continue
			case hyperLinkEnd:
				i += len([]rune(match[ANCHOR])) - 1
				builder.WriteString(formats.HyperlinkEnd)
				continue
			case empty:
				i += len([]rune(match[ANCHOR])) - 1
				continue
			}

			i = writeArchorOverride(match, background, i)
			continue
		}

		write(s)
	}

	// reset colors
	writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	currentColor.Pop()
}

func Len() int {
	return length
}

func String() (string, int) {
	defer func() {
		length = 0
		builder.Reset()

		isTransparent = false
		isInvisible = false
	}()

	return builder.String(), length
}

func writeEscapedAnsiString(text string) {
	if Plain {
		return
	}

	if len(formats.Escape) != 0 {
		text = fmt.Sprintf(formats.Escape, text)
	}

	builder.WriteString(text)
}

func write(s rune) {
	if isInvisible {
		return
	}

	if isHyperlink {
		builder.WriteRune(s)
		return
	}

	// UNSOLVABLE: When "Interactive" is true, the prompt length calculation in Bash/Zsh can be wrong, since the final string expansion is done by shells.
	length += runewidth.RuneWidth(s)
	// length += utf8.RuneCountInString(string(s))

	if !Interactive && !Plain {
		escaped, shouldEscape := formats.EscapeSequences[s]
		if shouldEscape {
			builder.WriteString(escaped)
			return
		}
	}

	builder.WriteRune(s)
}

func writeSegmentColors() {
	// use correct starting colors
	bg := backgroundColor
	fg := foregroundColor
	if !currentColor.Background().IsEmpty() {
		bg = currentColor.Background()
	}
	if !currentColor.Foreground().IsEmpty() {
		fg = currentColor.Foreground()
	}

	// ignore processing fully tranparent colors
	isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if isInvisible {
		return
	}

	switch {
	case fg.IsTransparent() && len(BackgroundColor) != 0:
		background := Colors.ToAnsi(BackgroundColor, false)
		writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		writeEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	case fg.IsTransparent() && !bg.IsEmpty():
		isTransparent = true
		writeEscapedAnsiString(fmt.Sprintf(transparentStart, bg))
	default:
		if !bg.IsEmpty() && !bg.IsTransparent() {
			writeEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}

		if !fg.IsEmpty() && !fg.IsTransparent() {
			writeEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	currentColor.Add(bg, fg)
}

func writeArchorOverride(match map[string]string, background color.Ansi, i int) int {
	position := i
	// check color reset first
	if match[ANCHOR] == resetStyle.AnchorEnd {
		return endColorOverride(position)
	}

	position += len([]rune(match[ANCHOR])) - 1

	for _, style := range knownStyles {
		if style.AnchorEnd == match[ANCHOR] {
			writeEscapedAnsiString(style.End)
			return position
		}
		if style.AnchorStart == match[ANCHOR] {
			writeEscapedAnsiString(style.Start)
			return position
		}
	}

	bgColor := color.Ansi(match[BG])
	fgColor := color.Ansi(match[FG])

	if fgColor.IsTransparent() && bgColor.IsEmpty() {
		bgColor = background
	}

	bg, fg := asAnsiColors(bgColor, fgColor)

	// ignore processing fully tranparent colors
	isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if isInvisible {
		return position
	}

	// make sure we have colors
	if fg.IsEmpty() {
		fg = foregroundColor
	}
	if bg.IsEmpty() {
		bg = backgroundColor
	}

	currentColor.Add(bg, fg)

	if currentColor.Foreground().IsTransparent() && len(BackgroundColor) != 0 {
		background := Colors.ToAnsi(BackgroundColor, false)
		writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		writeEscapedAnsiString(fmt.Sprintf(colorise, currentColor.Background().ToForeground()))
		return position
	}

	if currentColor.Foreground().IsTransparent() && !currentColor.Background().IsTransparent() {
		isTransparent = true
		writeEscapedAnsiString(fmt.Sprintf(transparentStart, currentColor.Background()))
		return position
	}

	if currentColor.Background() != backgroundColor {
		// end the colors in case we have a transparent background
		if currentColor.Background().IsTransparent() {
			writeEscapedAnsiString(backgroundEnd)
		} else {
			writeEscapedAnsiString(fmt.Sprintf(colorise, currentColor.Background()))
		}
	}

	if currentColor.Foreground() != foregroundColor {
		writeEscapedAnsiString(fmt.Sprintf(colorise, currentColor.Foreground()))
	}

	return position
}

func endColorOverride(position int) int {
	// make sure to reset the colors if needed
	position += len([]rune(resetStyle.AnchorEnd)) - 1

	// do not restore colors at the end of the string, we print it anyways
	if position == len(runes)-1 {
		currentColor.Pop()
		return position
	}

	// reset colors to previous when we have more than 1 in stack
	// as soon as we have  more than 1, we can pop the last one
	// and print the previous override as it wasn't ended yet
	if currentColor.Len() > 1 {
		fg := currentColor.Foreground()
		bg := currentColor.Background()

		currentColor.Pop()

		previousBg := currentColor.Background()
		previousFg := currentColor.Foreground()

		if isTransparent {
			writeEscapedAnsiString(transparentEnd)
		}

		if previousBg != bg {
			background := fmt.Sprintf(colorise, previousBg)
			if previousBg.IsClear() {
				background = backgroundStyle.End
			}

			writeEscapedAnsiString(background)
		}

		if previousFg != fg {
			writeEscapedAnsiString(fmt.Sprintf(colorise, previousFg))
		}

		return position
	}

	// pop the last colors from the stack
	defer currentColor.Pop()

	// do not reset when colors are identical
	if currentColor.Background() == backgroundColor && currentColor.Foreground() == foregroundColor {
		return position
	}

	if isTransparent {
		writeEscapedAnsiString(transparentEnd)
	}

	if backgroundColor.IsClear() {
		writeEscapedAnsiString(backgroundStyle.End)
	}

	if currentColor.Background() != backgroundColor && !backgroundColor.IsClear() {
		writeEscapedAnsiString(fmt.Sprintf(colorise, backgroundColor))
	}

	if (currentColor.Foreground() != foregroundColor || isTransparent) && !foregroundColor.IsClear() {
		writeEscapedAnsiString(fmt.Sprintf(colorise, foregroundColor))
	}

	isTransparent = false
	return position
}

func asAnsiColors(background, foreground color.Ansi) (color.Ansi, color.Ansi) {
	if len(background) == 0 {
		background = color.Background
	}

	if len(foreground) == 0 {
		foreground = color.Foreground
	}

	background = background.Resolve(CurrentColors, ParentColors)
	foreground = foreground.Resolve(CurrentColors, ParentColors)

	if bg, err := Colors.Resolve(background); err == nil {
		background = bg
	}

	if fg, err := Colors.Resolve(foreground); err == nil {
		foreground = fg
	}

	inverted := foreground == color.Transparent && len(background) != 0

	background = Colors.ToAnsi(background, !inverted)
	foreground = Colors.ToAnsi(foreground, false)

	return background, foreground
}

func trimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}
