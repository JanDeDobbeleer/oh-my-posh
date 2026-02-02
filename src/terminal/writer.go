package terminal

import (
	"fmt"
	"os"
	"strings"

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
)

// Package-level variables kept for CLI backward compatibility.
// CLI code paths set these directly; package-level functions sync them to defaultWriter.
var (
	BackgroundColor color.Ansi
	CurrentColors   *color.Set
	ParentColors    []*color.Set
	Colors          color.String

	Plain       bool
	Interactive bool

	Shell   string
	Program string

	formats *shell.Formats
)

// defaultWriter is used by package-level wrapper functions for CLI backward compat.
var defaultWriter = &Writer{}

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
	setProgress   = "\x1b]9;4;4;%d\x07"
	endProgress   = "\x1b]9;4;0;0\x07"

	WindowsTerminal = "Windows Terminal"
	Warp            = "WarpTerminal"
	ITerm           = "iTerm.app"
	AppleTerminal   = "Apple_Terminal"
	Unknown         = "Unknown"
)

// Writer holds all per-render mutable state for terminal output.
// Each render request should use its own Writer instance to avoid shared state.
type Writer struct {
	Colors          color.String
	formats         *shell.Formats
	CurrentColors   *color.Set
	foregroundColor color.Ansi
	backgroundColor color.Ansi
	BackgroundColor color.Ansi
	builder         strings.Builder
	currentColor    color.History
	runes           []rune
	ParentColors    []*color.Set
	length          int
	isTransparent   bool
	isInvisible     bool
	isHyperlink     bool
	Plain           bool
	Interactive     bool
}

// NewWriter creates a new Writer instance with shell-specific formats.
// Used by the daemon to create per-request writers.
func NewWriter(sh string) *Writer {
	return &Writer{
		formats: shell.GetFormats(sh),
	}
}

// syncDefaultWriter copies package-level config vars to defaultWriter
// so that package-level wrapper functions work correctly for CLI paths.
func syncDefaultWriter() {
	defaultWriter.BackgroundColor = BackgroundColor
	defaultWriter.CurrentColors = CurrentColors
	defaultWriter.ParentColors = ParentColors
	defaultWriter.Colors = Colors
	defaultWriter.Plain = Plain
	defaultWriter.Interactive = Interactive
	defaultWriter.formats = formats
}

// --- Package-level functions for CLI backward compat ---

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
	syncDefaultWriter()
	defaultWriter.SetColors(background, foreground)
	// Sync back to package-level for any code reading CurrentColors directly
	CurrentColors = defaultWriter.CurrentColors
}

func SetParentColors(background, foreground color.Ansi) {
	syncDefaultWriter()
	defaultWriter.SetParentColors(background, foreground)
	ParentColors = defaultWriter.ParentColors
}

func Write(background, foreground color.Ansi, txt string) {
	syncDefaultWriter()
	defaultWriter.Write(background, foreground, txt)
}

func Len() int {
	return defaultWriter.Len()
}

func String() (string, int) {
	return defaultWriter.String()
}

// --- Package-level functions that only read Shell/Program/formats (no builder state) ---

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

// --- Writer methods ---

func (w *Writer) SetColors(background, foreground color.Ansi) {
	w.CurrentColors = &color.Set{
		Background: background,
		Foreground: foreground,
	}
}

func (w *Writer) SetParentColors(background, foreground color.Ansi) {
	if w.ParentColors == nil {
		w.ParentColors = make([]*color.Set, 0)
	}

	w.ParentColors = append([]*color.Set{{
		Background: background,
		Foreground: foreground,
	}}, w.ParentColors...)
}

func (w *Writer) Write(background, foreground color.Ansi, txt string) {
	if txt == "" {
		return
	}

	w.backgroundColor, w.foregroundColor = w.asAnsiColors(background, foreground)

	// default to white foreground
	if w.foregroundColor.IsEmpty() {
		w.foregroundColor = w.Colors.ToAnsi("white", false)
	}

	// validate if we start with a color override
	match := regex.FindNamedRegexMatch(AnchorRegex, txt)
	if len(match) != 0 && match[ANCHOR] != hyperLinkStart {
		colorOverride := true
		for _, style := range knownStyles {
			if match[ANCHOR] != style.AnchorStart {
				continue
			}

			w.writeEscapedAnsiString(style.Start)
			colorOverride = false
		}

		if colorOverride {
			w.currentColor.Add(w.asAnsiColors(color.Ansi(match[BG]), color.Ansi(match[FG])))
		}
	}

	w.writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match[ANCHOR] == hyperLinkStart {
		w.isHyperlink = true
		w.builder.WriteString(w.formats.HyperlinkStart)
	}

	txt = txt[len(match[ANCHOR]):]
	w.runes = []rune(txt)
	hyperlinkTextPosition := 0

	for i := 0; i < len(w.runes); i++ {
		s := w.runes[i]
		// ignore everything which isn't overriding
		if s != '<' {
			w.write(s)
			continue
		}

		// color/end overrides first
		txt = string(w.runes[i:])
		match = regex.FindNamedRegexMatch(AnchorRegex, txt)
		if len(match) > 0 {
			// check for hyperlinks first
			switch match[ANCHOR] {
			case hyperLinkStart:
				w.isHyperlink = true
				i += len([]rune(match[ANCHOR])) - 1
				w.builder.WriteString(w.formats.HyperlinkStart)
				continue
			case hyperLinkText:
				w.isHyperlink = false
				i += len([]rune(match[ANCHOR])) - 1
				hyperlinkTextPosition = i
				w.builder.WriteString(w.formats.HyperlinkCenter)
				continue
			case hyperLinkTextEnd:
				// this implies there's no text in the hyperlink
				if hyperlinkTextPosition+1 == i {
					w.builder.WriteString("link")
					w.length += 4
				}
				i += len([]rune(match[ANCHOR])) - 1
				continue
			case hyperLinkEnd:
				i += len([]rune(match[ANCHOR])) - 1
				w.builder.WriteString(w.formats.HyperlinkEnd)
				continue
			case empty:
				i += len([]rune(match[ANCHOR])) - 1
				continue
			}

			i = w.writeAnchorOverride(match, background, i)
			continue
		}

		w.write(s)
	}

	// reset colors
	w.writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	w.currentColor.Pop()
}

func (w *Writer) Len() int {
	return w.length
}

func (w *Writer) String() (string, int) {
	defer func() {
		w.length = 0
		w.builder.Reset()

		w.isTransparent = false
		w.isInvisible = false
	}()

	return w.builder.String(), w.length
}

func (w *Writer) writeEscapedAnsiString(txt string) {
	if w.Plain {
		return
	}

	if len(w.formats.Escape) != 0 {
		txt = fmt.Sprintf(w.formats.Escape, txt)
	}

	w.builder.WriteString(txt)
}

func (w *Writer) write(s rune) {
	if w.isInvisible {
		return
	}

	if w.isHyperlink {
		w.builder.WriteRune(s)
		return
	}

	// UNSOLVABLE: When "Interactive" is true, the prompt length calculation in Bash/Zsh can be wrong, since the final string expansion is done by shells.
	w.length += runewidth.RuneWidth(s)

	if !w.Interactive && !w.Plain {
		escaped, shouldEscape := w.formats.EscapeSequences[s]
		if shouldEscape {
			w.builder.WriteString(escaped)
			return
		}
	}

	w.builder.WriteRune(s)
}

func (w *Writer) writeSegmentColors() {
	// use correct starting colors
	bg := w.backgroundColor
	fg := w.foregroundColor
	if !w.currentColor.Background().IsEmpty() {
		bg = w.currentColor.Background()
	}
	if !w.currentColor.Foreground().IsEmpty() {
		fg = w.currentColor.Foreground()
	}

	// ignore processing fully transparent colors
	w.isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if w.isInvisible {
		return
	}

	switch {
	case fg.IsTransparent() && len(w.BackgroundColor) != 0:
		background := w.Colors.ToAnsi(w.BackgroundColor, false)
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	case fg.IsTransparent() && !bg.IsEmpty():
		w.isTransparent = true
		w.writeEscapedAnsiString(fmt.Sprintf(transparentStart, bg))
	default:
		if !bg.IsEmpty() && !bg.IsTransparent() {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}

		if !fg.IsEmpty() && !fg.IsTransparent() {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	w.currentColor.Add(bg, fg)
}

func (w *Writer) writeAnchorOverride(match map[string]string, background color.Ansi, i int) int {
	position := i
	// check color reset first
	if match[ANCHOR] == resetStyle.AnchorEnd {
		return w.endColorOverride(position)
	}

	position += len([]rune(match[ANCHOR])) - 1

	for _, style := range knownStyles {
		if style.AnchorEnd == match[ANCHOR] {
			w.writeEscapedAnsiString(style.End)
			return position
		}
		if style.AnchorStart == match[ANCHOR] {
			w.writeEscapedAnsiString(style.Start)
			return position
		}
	}

	bgColor := color.Ansi(match[BG])
	fgColor := color.Ansi(match[FG])

	if fgColor.IsTransparent() && bgColor.IsEmpty() {
		bgColor = background
	}

	bg, fg := w.asAnsiColors(bgColor, fgColor)

	// ignore processing fully transparent colors
	w.isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if w.isInvisible {
		return position
	}

	// make sure we have colors
	if fg.IsEmpty() {
		fg = w.foregroundColor
	}
	if bg.IsEmpty() {
		bg = w.backgroundColor
	}

	w.currentColor.Add(bg, fg)

	if w.currentColor.Foreground().IsTransparent() && len(w.BackgroundColor) != 0 {
		background := w.Colors.ToAnsi(w.BackgroundColor, false)
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, w.currentColor.Background().ToForeground()))
		return position
	}

	if w.currentColor.Foreground().IsTransparent() && !w.currentColor.Background().IsTransparent() {
		w.isTransparent = true
		w.writeEscapedAnsiString(fmt.Sprintf(transparentStart, w.currentColor.Background()))
		return position
	}

	if w.currentColor.Background() != w.backgroundColor {
		// end the colors in case we have a transparent background
		if w.currentColor.Background().IsTransparent() {
			w.writeEscapedAnsiString(backgroundEnd)
		} else {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, w.currentColor.Background()))
		}
	}

	if w.currentColor.Foreground() != w.foregroundColor {
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, w.currentColor.Foreground()))
	}

	return position
}

func (w *Writer) endColorOverride(position int) int {
	// make sure to reset the colors if needed
	position += len([]rune(resetStyle.AnchorEnd)) - 1

	// do not restore colors at the end of the string, we print it anyways
	if position == len(w.runes)-1 {
		w.currentColor.Pop()
		return position
	}

	// reset colors to previous when we have more than 1 in stack
	// as soon as we have  more than 1, we can pop the last one
	// and print the previous override as it wasn't ended yet
	if w.currentColor.Len() > 1 {
		fg := w.currentColor.Foreground()
		bg := w.currentColor.Background()

		w.currentColor.Pop()

		previousBg := w.currentColor.Background()
		previousFg := w.currentColor.Foreground()

		if w.isTransparent {
			w.writeEscapedAnsiString(transparentEnd)
		}

		if previousBg != bg {
			background := fmt.Sprintf(colorise, previousBg)
			if previousBg.IsClear() {
				background = backgroundStyle.End
			}

			w.writeEscapedAnsiString(background)
		}

		if previousFg != fg {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, previousFg))
		}

		return position
	}

	// pop the last colors from the stack
	defer w.currentColor.Pop()

	// do not reset when colors are identical
	if w.currentColor.Background() == w.backgroundColor && w.currentColor.Foreground() == w.foregroundColor {
		return position
	}

	if w.isTransparent {
		w.writeEscapedAnsiString(transparentEnd)
	}

	if w.backgroundColor.IsClear() {
		w.writeEscapedAnsiString(backgroundStyle.End)
	}

	if w.currentColor.Background() != w.backgroundColor && !w.backgroundColor.IsClear() {
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, w.backgroundColor))
	}

	if (w.currentColor.Foreground() != w.foregroundColor || w.isTransparent) && !w.foregroundColor.IsClear() {
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, w.foregroundColor))
	}

	w.isTransparent = false
	return position
}

func (w *Writer) asAnsiColors(background, foreground color.Ansi) (color.Ansi, color.Ansi) {
	if background == "" {
		background = color.Background
	}

	if foreground == "" {
		foreground = color.Foreground
	}

	background = background.Resolve(w.CurrentColors, w.ParentColors)
	foreground = foreground.Resolve(w.CurrentColors, w.ParentColors)

	if bg, err := w.Colors.Resolve(background); err == nil {
		background = bg
	}

	if fg, err := w.Colors.Resolve(foreground); err == nil {
		foreground = fg
	}

	inverted := foreground == color.Transparent && len(background) != 0

	background = w.Colors.ToAnsi(background, !inverted)
	foreground = w.Colors.ToAnsi(foreground, false)

	return background, foreground
}

func trimAnsi(txt string) string {
	if txt == "" || !strings.Contains(txt, "\x1b") {
		return txt
	}
	return regex.ReplaceAllString(AnsiRegex, txt, "")
}
