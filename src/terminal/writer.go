package terminal

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/mattn/go-runewidth"
)

func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}

type Colors struct {
	Background string `json:"background" toml:"background"`
	Foreground string `json:"foreground" toml:"foreground"`
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

type shellFormats struct {
	escape     string
	left       string
	linechange string
	clearBelow string
	clearLine  string

	title string

	saveCursorPosition    string
	restoreCursorPosition string

	osc99 string
	osc7  string
	osc51 string

	escapeSequences map[rune]rune

	hyperlinkStart  string
	hyperlinkCenter string
	hyperlinkEnd    string

	iTermPromptMark string
	iTermCurrentDir string
	iTermRemoteHost string
}

const (
	// Transparent implies a transparent color
	Transparent = "transparent"
	// Accent is the OS accent color
	Accent = "accent"
	// ParentBackground takes the previous segment's background color
	ParentBackground = "parentBackground"
	// ParentForeground takes the previous segment's color
	ParentForeground = "parentForeground"
	// Background takes the current segment's background color
	Background = "background"
	// Foreground takes the current segment's foreground color
	Foreground = "foreground"

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
)

// Writer writes colorized ANSI strings
type Writer struct {
	BackgroundColor string
	CurrentColors   *Colors
	ParentColors    []*Colors
	AnsiColors      ColorString

	Plain       bool
	TrueColor   bool
	Interactive bool

	builder strings.Builder
	length  int

	foregroundColor Color
	backgroundColor Color
	currentColor    ColorHistory
	runes           []rune

	isTransparent bool
	isInvisible   bool
	isHyperlink   bool

	lastRune rune

	shellName string

	formats *shellFormats
}

func (w *Writer) Init(shellName string) {
	w.shellName = shellName

	switch w.shellName {
	case shell.BASH:
		w.formats = &shellFormats{
			escape:                "\\[%s\\]",
			linechange:            "\\[\x1b[%d%s\\]",
			left:                  "\\[\x1b[%dD\\]",
			clearBelow:            "\\[\x1b[0J\\]",
			clearLine:             "\\[\x1b[K\\]",
			saveCursorPosition:    "\\[\x1b7\\]",
			restoreCursorPosition: "\\[\x1b8\\]",
			title:                 "\\[\x1b]0;%s\007\\]",
			hyperlinkStart:        "\\[\x1b]8;;",
			hyperlinkCenter:       "\x1b\\\\\\]",
			hyperlinkEnd:          "\\[\x1b]8;;\x1b\\\\\\]",
			osc99:                 "\\[\x1b]9;9;%s\x1b\\\\\\]",
			osc7:                  "\\[\x1b]7;file://%s/%s\x1b\\\\\\]",
			osc51:                 "\\[\x1b]51;A;%s@%s:%s\x1b\\\\\\]",
			iTermPromptMark:       "\\[$(iterm2_prompt_mark)\\]",
			iTermCurrentDir:       "\\[\x1b]1337;CurrentDir=%s\x07\\]",
			iTermRemoteHost:       "\\[\x1b]1337;RemoteHost=%s@%s\x07\\]",
			escapeSequences: map[rune]rune{
				96: 92, // backtick
				92: 92, // backslash
			},
		}
	case shell.ZSH, shell.TCSH:
		w.formats = &shellFormats{
			escape:                "%%{%s%%}",
			linechange:            "%%{\x1b[%d%s%%}",
			left:                  "%%{\x1b[%dD%%}",
			clearBelow:            "%{\x1b[0J%}",
			clearLine:             "%{\x1b[K%}",
			saveCursorPosition:    "%{\x1b7%}",
			restoreCursorPosition: "%{\x1b8%}",
			title:                 "%%{\x1b]0;%s\007%%}",
			hyperlinkStart:        "%{\x1b]8;;",
			hyperlinkCenter:       "\x1b\\%}",
			hyperlinkEnd:          "%{\x1b]8;;\x1b\\%}",
			osc99:                 "%%{\x1b]9;9;%s\x1b\\%%}",
			osc7:                  "%%{\x1b]7;file://%s/%s\x1b\\%%}",
			osc51:                 "%%{\x1b]51;A%s@%s:%s\x1b\\%%}",
			iTermPromptMark:       "%{$(iterm2_prompt_mark)%}",
			iTermCurrentDir:       "%%{\x1b]1337;CurrentDir=%s\x07%%}",
			iTermRemoteHost:       "%%{\x1b]1337;RemoteHost=%s@%s\x07%%}",
		}
	default:
		w.formats = &shellFormats{
			escape:                "%s",
			linechange:            "\x1b[%d%s",
			left:                  "\x1b[%dD",
			clearBelow:            "\x1b[0J",
			clearLine:             "\x1b[K",
			saveCursorPosition:    "\x1b7",
			restoreCursorPosition: "\x1b8",
			title:                 "\x1b]0;%s\007",
			// when in fish on Linux, it seems hyperlinks ending with \\ print a \
			// unlike on macOS. However, this is a fish bug, so do not try to fix it here:
			// https://github.com/JanDeDobbeleer/oh-my-posh/pull/3288#issuecomment-1369137068
			hyperlinkStart:  "\x1b]8;;",
			hyperlinkCenter: "\x1b\\",
			hyperlinkEnd:    "\x1b]8;;\x1b\\",
			osc99:           "\x1b]9;9;%s\x1b\\",
			osc7:            "\x1b]7;file://%s/%s\x1b\\",
			osc51:           "\x1b]51;A%s@%s:%s\x1b\\",
			iTermCurrentDir: "\x1b]1337;CurrentDir=%s\x07",
			iTermRemoteHost: "\x1b]1337;RemoteHost=%s@%s\x07",
		}
	}

	if shellName == shell.ZSH {
		w.formats.escapeSequences = map[rune]rune{
			96: 92, // backtick
			37: 37, // %
		}
	}
}

func (w *Writer) SetColors(background, foreground string) {
	w.CurrentColors = &Colors{
		Background: background,
		Foreground: foreground,
	}
}

func (w *Writer) SetParentColors(background, foreground string) {
	if w.ParentColors == nil {
		w.ParentColors = make([]*Colors, 0)
	}

	w.ParentColors = append([]*Colors{{
		Background: background,
		Foreground: foreground,
	}}, w.ParentColors...)
}

func (w *Writer) ChangeLine(numberOfLines int) string {
	if w.Plain {
		return ""
	}

	position := "B"

	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}

	return fmt.Sprintf(w.formats.linechange, numberOfLines, position)
}

func (w *Writer) ConsolePwd(pwdType, userName, hostName, pwd string) string {
	if w.Plain {
		return ""
	}

	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}

	switch pwdType {
	case OSC7:
		return fmt.Sprintf(w.formats.osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(w.formats.osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(w.formats.osc99, pwd)
	}
}

func (w *Writer) ClearAfter() string {
	if w.Plain {
		return ""
	}

	return w.formats.clearLine + w.formats.clearBelow
}

func (w *Writer) FormatTitle(title string) string {
	title = w.trimAnsi(title)

	if w.Plain {
		return title
	}

	// we have to do this to prevent bash/zsh from misidentifying escape sequences
	switch w.shellName {
	case shell.BASH:
		title = strings.NewReplacer("`", "\\`", `\`, `\\`).Replace(title)
	case shell.ZSH:
		title = strings.NewReplacer("`", "\\`", `%`, `%%`).Replace(title)
	case shell.ELVISH, shell.XONSH:
		// these shells don't support setting the title
		return ""
	}

	return fmt.Sprintf(w.formats.title, title)
}

func (w *Writer) EscapeText(text string) string {
	return fmt.Sprintf(w.formats.escape, text)
}

func (w *Writer) SaveCursorPosition() string {
	return w.formats.saveCursorPosition
}

func (w *Writer) RestoreCursorPosition() string {
	return w.formats.restoreCursorPosition
}

func (w *Writer) PromptStart() string {
	return fmt.Sprintf(w.formats.escape, "\x1b]133;A\007")
}

func (w *Writer) CommandStart() string {
	return fmt.Sprintf(w.formats.escape, "\x1b]133;B\007")
}

func (w *Writer) CommandFinished(code int, ignore bool) string {
	if ignore {
		return fmt.Sprintf(w.formats.escape, "\x1b]133;D\007")
	}

	mark := fmt.Sprintf("\x1b]133;D;%d\007", code)

	return fmt.Sprintf(w.formats.escape, mark)
}

func (w *Writer) LineBreak() string {
	cr := fmt.Sprintf(w.formats.left, 1000)
	lf := fmt.Sprintf(w.formats.linechange, 1, "B")
	return cr + lf
}

func (w *Writer) Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	w.backgroundColor, w.foregroundColor = w.asAnsiColors(background, foreground)

	// default to white foreground
	if w.foregroundColor.IsEmpty() {
		w.foregroundColor = w.AnsiColors.ToColor("white", false, w.TrueColor)
	}

	// validate if we start with a color override
	match := regex.FindNamedRegexMatch(AnchorRegex, text)
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
			w.currentColor.Add(w.asAnsiColors(match[BG], match[FG]))
		}
	}

	w.writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match[ANCHOR] == hyperLinkStart {
		w.isHyperlink = true
		w.builder.WriteString(w.formats.hyperlinkStart)
	}

	text = text[len(match[ANCHOR]):]
	w.runes = []rune(text)
	hyperlinkTextPosition := 0

	for i := 0; i < len(w.runes); i++ {
		s := w.runes[i]
		// ignore everything which isn't overriding
		if s != '<' {
			w.write(s)
			continue
		}

		// color/end overrides first
		text = string(w.runes[i:])
		match = regex.FindNamedRegexMatch(AnchorRegex, text)
		if len(match) > 0 {
			// check for hyperlinks first
			switch match[ANCHOR] {
			case hyperLinkStart:
				w.isHyperlink = true
				i += len([]rune(match[ANCHOR])) - 1
				w.builder.WriteString(w.formats.hyperlinkStart)
				continue
			case hyperLinkText:
				w.isHyperlink = false
				i += len([]rune(match[ANCHOR])) - 1
				hyperlinkTextPosition = i
				w.builder.WriteString(w.formats.hyperlinkCenter)
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
				w.builder.WriteString(w.formats.hyperlinkEnd)
				continue
			}

			i = w.writeArchorOverride(match, background, i)
			continue
		}

		w.length += runewidth.RuneWidth(s)
		w.write(s)
	}

	// reset colors
	w.writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	w.currentColor.Pop()
}

func (w *Writer) String() (string, int) {
	defer func() {
		w.length = 0
		w.builder.Reset()
	}()

	return w.builder.String(), w.length
}

func (w *Writer) writeEscapedAnsiString(text string) {
	if w.Plain {
		return
	}

	if len(w.formats.escape) != 0 {
		text = fmt.Sprintf(w.formats.escape, text)
	}

	w.builder.WriteString(text)
}

func (w *Writer) getAnsiFromColorString(colorString string, isBackground bool) Color {
	return w.AnsiColors.ToColor(colorString, isBackground, w.TrueColor)
}

func (w *Writer) write(s rune) {
	if w.isInvisible {
		return
	}

	if w.isHyperlink {
		w.builder.WriteRune(s)
		return
	}

	if !w.Interactive {
		for special, escape := range w.formats.escapeSequences {
			if s == special && w.lastRune != escape {
				w.builder.WriteRune(escape)
			}
		}
	}

	w.length += runewidth.RuneWidth(s)
	w.lastRune = s
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

	// ignore processing fully tranparent colors
	w.isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if w.isInvisible {
		return
	}

	if fg.IsTransparent() && len(w.BackgroundColor) != 0 { //nolint: gocritic
		background := w.getAnsiFromColorString(w.BackgroundColor, false)
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		w.writeEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	} else if fg.IsTransparent() && !bg.IsEmpty() {
		w.isTransparent = true
		w.writeEscapedAnsiString(fmt.Sprintf(transparentStart, bg))
	} else {
		if !bg.IsEmpty() && !bg.IsTransparent() {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}
		if !fg.IsEmpty() {
			w.writeEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	w.currentColor.Add(bg, fg)
}

func (w *Writer) writeArchorOverride(match map[string]string, background string, i int) int {
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

	if match[FG] == Transparent && len(match[BG]) == 0 {
		match[BG] = background
	}

	bg, fg := w.asAnsiColors(match[BG], match[FG])

	// ignore processing fully tranparent colors
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
		background := w.getAnsiFromColorString(w.BackgroundColor, false)
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

func (w *Writer) asAnsiColors(background, foreground string) (Color, Color) {
	if len(background) == 0 {
		background = Background
	}

	if len(foreground) == 0 {
		foreground = Foreground
	}

	background = w.expandKeyword(background)
	foreground = w.expandKeyword(foreground)

	inverted := foreground == Transparent && len(background) != 0

	backgroundAnsi := w.getAnsiFromColorString(background, !inverted)
	foregroundAnsi := w.getAnsiFromColorString(foreground, false)

	return backgroundAnsi, foregroundAnsi
}

func (w *Writer) isKeyword(color string) bool {
	switch color {
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func (w *Writer) expandKeyword(keyword string) string {
	resolveParentColor := func(keyword string) string {
		for _, color := range w.ParentColors {
			if color == nil {
				return Transparent
			}
			switch keyword {
			case ParentBackground:
				keyword = color.Background
			case ParentForeground:
				keyword = color.Foreground
			default:
				if len(keyword) == 0 {
					return Transparent
				}
				return keyword
			}
		}

		if len(keyword) == 0 {
			return Transparent
		}

		return keyword
	}

	resolveKeyword := func(keyword string) string {
		switch {
		case keyword == Background && w.CurrentColors != nil:
			return w.CurrentColors.Background
		case keyword == Foreground && w.CurrentColors != nil:
			return w.CurrentColors.Foreground
		case (keyword == ParentBackground || keyword == ParentForeground) && w.ParentColors != nil:
			return resolveParentColor(keyword)
		default:
			return Transparent
		}
	}

	for ok := w.isKeyword(keyword); ok; ok = w.isKeyword(keyword) {
		resolved := resolveKeyword(keyword)
		if resolved == keyword {
			break
		}

		keyword = resolved
	}

	return keyword
}

func (w *Writer) trimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}
