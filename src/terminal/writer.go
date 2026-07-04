package terminal

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
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
	textLen         int

	isTransparent bool
	isInvisible   bool
	isHyperlink   bool

	Shell   string
	Program string

	formats *shell.Formats

	// escapePrefix/escapeSuffix are formats.Escape ("...%s...") split around its
	// single %s placeholder, precomputed once so writeEscapedAnsiString can
	// concatenate via the builder instead of allocating through fmt.Sprintf.
	escapePrefix string
	escapeSuffix string
)

const (
	AnchorRegex = `^(?P<ANCHOR><(?P<FG>[^,<>]+)?,?(?P<BG>[^<>]+)?>)`

	// colorisePrefix/coloriseSuffix and transparentStartPrefix/transparentStartSuffix
	// are the fixed parts of the colorise ("\x1b[%sm") and transparentStart
	// ("\x1b[0m\x1b[%s;49m\x1b[7m") formats, split around their single %s
	// placeholder so callers can write them directly via the builder instead
	// of allocating through fmt.Sprintf.
	colorisePrefix = "\x1b["
	coloriseSuffix = "m"

	transparentStartPrefix = "\x1b[0m\x1b["
	transparentStartSuffix = ";49m\x1b[7m"

	transparentEnd = "\x1b[27m"
	backgroundEnd  = "\x1b[49m"

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
	setProgress   = "\x1b]9;4;4;%d\x07"
	endProgress   = "\x1b]9;4;0;0\x07"

	WindowsTerminal = "Windows Terminal"
	Warp            = "WarpTerminal"
	ITerm           = "iTerm.app"
	AppleTerminal   = "Apple_Terminal"
	Unknown         = "Unknown"
)

// anchorMatch describes a single `<...>` anchor token found while scanning
// segment text, mirroring the named groups of AnchorRegex without allocating
// a map or invoking the regexp engine.
type anchorMatch struct {
	Anchor string
	FG     string
	BG     string
	ok     bool
}

// scanAnchor looks for an AnchorRegex-shaped token at the start of txt, i.e.
// `<` + zero or more non `<>` characters + `>`, with the inner text optionally
// split on the first comma into FG (before) and BG (after, which may itself
// contain further commas). It operates on the zero-copy slice txt[i:] and
// performs no allocations.
func scanAnchor(txt string) anchorMatch {
	if len(txt) == 0 || txt[0] != '<' {
		return anchorMatch{}
	}

	end := strings.IndexByte(txt, '>')
	if end < 0 {
		return anchorMatch{}
	}

	inner := txt[1:end]
	if strings.IndexByte(inner, '<') >= 0 {
		return anchorMatch{}
	}

	fg := inner
	bg := ""

	if before, after, ok := strings.Cut(inner, ","); ok {
		fg = before
		bg = after
	}

	return anchorMatch{
		Anchor: txt[:end+1],
		FG:     fg,
		BG:     bg,
		ok:     true,
	}
}

func Init(sh string) {
	Shell = sh
	Program = getTerminalName()

	log.Debug("terminal program:", Program)
	log.Debug("terminal shell:", Shell)

	color.TrueColor = Program != AppleTerminal

	formats = shell.GetFormats(Shell)

	escapePrefix, escapeSuffix = "", ""
	if before, after, found := strings.Cut(formats.Escape, "%s"); found {
		// formats.Escape is a fmt.Sprintf format string, so any literal "%"
		// in the surrounding text is escaped as "%%" (e.g. zsh's "%%{%s%%}").
		// Unescape it now since we no longer route through fmt.Sprintf.
		escapePrefix = strings.ReplaceAll(before, "%%", "%")
		escapeSuffix = strings.ReplaceAll(after, "%%", "%")
	}
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
	case shell.ELVISH, shell.XONSH:
		return ""
	case shell.BASH, shell.ZSH:
		title = trimAnsi(title)

		sb := text.NewBuilder()

		// We have to do this to prevent the shell from misidentifying escape sequences.
		for _, char := range title {
			escaped, shouldEscape := formats.EscapeSequences[char]
			if shouldEscape {
				sb.WriteString(escaped)
				continue
			}

			sb.WriteRune(char)
		}

		return fmt.Sprintf(formats.Title, sb.String())
	default:
		return fmt.Sprintf(formats.Title, trimAnsi(title))
	}
}

func EscapeText(txt string) string {
	return fmt.Sprintf(formats.Escape, txt)
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

func SetProgress(percentage int) string {
	if Program != WindowsTerminal {
		return ""
	}

	return fmt.Sprintf(setProgress, percentage)
}

func StopProgress() string {
	if Program != WindowsTerminal {
		return ""
	}

	return endProgress
}

func Write(background, foreground color.Ansi, txt string) {
	if txt == "" {
		return
	}

	backgroundColor, foregroundColor = asAnsiColors(background, foreground)

	// default to white foreground
	if foregroundColor.IsEmpty() {
		foregroundColor = Colors.ToAnsi("white", false)
	}

	// validate if we start with a color override
	match := scanAnchor(txt)
	if match.ok && match.Anchor != hyperLinkStart {
		colorOverride := true
		for _, style := range knownStyles {
			if match.Anchor != style.AnchorStart {
				continue
			}

			writeEscapedAnsiString(style.Start)
			colorOverride = false
		}

		if colorOverride {
			currentColor.Add(asAnsiColors(color.Ansi(match.BG), color.Ansi(match.FG)))
		}
	}

	writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match.ok && match.Anchor == hyperLinkStart {
		isHyperlink = true
		builder.WriteString(formats.HyperlinkStart)
	}

	txt = txt[len(match.Anchor):]
	textLen = len(txt)
	hyperlinkTextPosition := 0

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		// ignore everything which isn't overriding
		if s != '<' {
			write(s)
			i += size
			continue
		}

		// color/end overrides first
		match = scanAnchor(txt[i:])
		if match.ok {
			// check for hyperlinks first
			switch match.Anchor {
			case hyperLinkStart:
				isHyperlink = true
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkStart)
				continue
			case hyperLinkText:
				isHyperlink = false
				i += len(match.Anchor)
				hyperlinkTextPosition = i
				builder.WriteString(formats.HyperlinkCenter)
				continue
			case hyperLinkTextEnd:
				// this implies there's no text in the hyperlink
				if hyperlinkTextPosition == i {
					builder.WriteString("link")
					length += 4
				}
				i += len(match.Anchor)
				continue
			case hyperLinkEnd:
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkEnd)
				continue
			case empty:
				i += len(match.Anchor)
				continue
			}

			i = writeAnchorOverride(match, background, i)
			continue
		}

		write(s)
		i += size
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

func writeEscapedAnsiString(txt string) {
	if Plain {
		return
	}

	if len(escapePrefix) != 0 {
		builder.WriteString(escapePrefix)
	}

	builder.WriteString(txt)

	if len(escapeSuffix) != 0 {
		builder.WriteString(escapeSuffix)
	}
}

// writeEscapedAnsiParts writes prefix+payload+suffix wrapped in the shell escape
// sequence, avoiding the intermediate string concatenation that a
// fmt.Sprintf(colorise/transparentStart, ...) call would otherwise require.
func writeEscapedAnsiParts(prefix string, payload color.Ansi, suffix string) {
	if Plain {
		return
	}

	if len(escapePrefix) != 0 {
		builder.WriteString(escapePrefix)
	}

	builder.WriteString(prefix)
	builder.WriteString(payload.String())
	builder.WriteString(suffix)

	if len(escapeSuffix) != 0 {
		builder.WriteString(escapeSuffix)
	}
}

// writeColorise writes the equivalent of fmt.Sprintf(colorise, c) wrapped in
// the shell escape sequence, without allocating an intermediate string.
func writeColorise(c color.Ansi) {
	writeEscapedAnsiParts(colorisePrefix, c, coloriseSuffix)
}

// writeTransparentStart writes the equivalent of fmt.Sprintf(transparentStart, c)
// wrapped in the shell escape sequence, without allocating an intermediate string.
func writeTransparentStart(c color.Ansi) {
	writeEscapedAnsiParts(transparentStartPrefix, c, transparentStartSuffix)
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

	// ignore processing fully transparent colors
	isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if isInvisible {
		return
	}

	switch {
	case fg.IsTransparent() && len(BackgroundColor) != 0:
		background := Colors.ToAnsi(BackgroundColor, false)
		writeColorise(background)
		writeColorise(bg.ToForeground())
	case fg.IsTransparent() && !bg.IsEmpty():
		isTransparent = true
		writeTransparentStart(bg)
	default:
		if !bg.IsEmpty() && !bg.IsTransparent() {
			writeColorise(bg)
		}

		if !fg.IsEmpty() && !fg.IsTransparent() {
			writeColorise(fg)
		}
	}

	// set current colors
	currentColor.Add(bg, fg)
}

func writeAnchorOverride(match anchorMatch, background color.Ansi, i int) int {
	position := i
	// check color reset first
	if match.Anchor == resetStyle.AnchorEnd {
		return endColorOverride(position)
	}

	position += len(match.Anchor)

	for _, style := range knownStyles {
		if style.AnchorEnd == match.Anchor {
			writeEscapedAnsiString(style.End)
			return position
		}
		if style.AnchorStart == match.Anchor {
			writeEscapedAnsiString(style.Start)
			return position
		}
	}

	bgColor := color.Ansi(match.BG)
	fgColor := color.Ansi(match.FG)

	if fgColor.IsTransparent() && bgColor.IsEmpty() {
		bgColor = background
	}

	bg, fg := asAnsiColors(bgColor, fgColor)

	// ignore processing fully transparent colors
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
		writeColorise(background)
		writeColorise(currentColor.Background().ToForeground())
		return position
	}

	if currentColor.Foreground().IsTransparent() && !currentColor.Background().IsTransparent() {
		isTransparent = true
		writeTransparentStart(currentColor.Background())
		return position
	}

	if currentColor.Background() != backgroundColor {
		// end the colors in case we have a transparent background
		if currentColor.Background().IsTransparent() {
			writeEscapedAnsiString(backgroundEnd)
		} else {
			writeColorise(currentColor.Background())
		}
	}

	if currentColor.Foreground() != foregroundColor {
		writeColorise(currentColor.Foreground())
	}

	return position
}

func endColorOverride(position int) int {
	// make sure to reset the colors if needed
	position += len(resetStyle.AnchorEnd)

	// do not restore colors at the end of the string, we print it anyways
	if position == textLen {
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
			if previousBg.IsClear() {
				writeEscapedAnsiString(backgroundStyle.End)
			} else {
				writeColorise(previousBg)
			}
		}

		if previousFg != fg {
			writeColorise(previousFg)
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
		writeColorise(backgroundColor)
	}

	if (currentColor.Foreground() != foregroundColor || isTransparent) && !foregroundColor.IsClear() {
		writeColorise(foregroundColor)
	}

	isTransparent = false
	return position
}

func asAnsiColors(background, foreground color.Ansi) (color.Ansi, color.Ansi) {
	if background == "" {
		background = color.Background
	}

	if foreground == "" {
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

func trimAnsi(txt string) string {
	if txt == "" || !strings.Contains(txt, "\x1b") {
		return txt
	}
	return regex.ReplaceAllString(AnsiRegex, txt, "")
}
