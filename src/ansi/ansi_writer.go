package ansi

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/regex"
	"github.com/jandedobbeleer/oh-my-posh/shell"
	"github.com/mattn/go-runewidth"
)

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
	colorStyle = &style{AnchorStart: "COLOR", AnchorEnd: `</>`, End: "\x1b[0m"}
)

type style struct {
	AnchorStart string
	AnchorEnd   string
	Start       string
	End         string
}

type cachedColor struct {
	Background string
	Foreground string
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

	anchorRegex = `^(?P<ANCHOR><(?P<FG>[^,>]+)?,?(?P<BG>[^>]+)?>)`
	colorise    = "\x1b[%sm"
	transparent = "\x1b[%s;49m\x1b[7m"

	AnsiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	OSC99 string = "osc99"
	OSC7  string = "osc7"
	OSC51 string = "osc51"
)

// Writer writes colorized ANSI strings
type Writer struct {
	TerminalBackground string
	Colors             *cachedColor
	ParentColors       []*cachedColor
	AnsiColors         Colors
	Plain              bool

	builder strings.Builder
	length  int

	foreground        Color
	background        Color
	currentForeground Color
	currentBackground Color
	runes             []rune

	shell                 string
	format                string
	left                  string
	right                 string
	title                 string
	linechange            string
	clearBelow            string
	clearLine             string
	saveCursorPosition    string
	restoreCursorPosition string
	escapeLeft            string
	escapeRight           string
	hyperlink             string
	hyperlinkRegex        string
	osc99                 string
	osc7                  string
	osc51                 string
}

func (w *Writer) Init(shellName string) {
	w.shell = shellName
	switch w.shell {
	case shell.BASH:
		w.format = "\\[%s\\]"
		w.linechange = "\\[\x1b[%d%s\\]"
		w.right = "\\[\x1b[%dC\\]"
		w.left = "\\[\x1b[%dD\\]"
		w.clearBelow = "\\[\x1b[0J\\]"
		w.clearLine = "\\[\x1b[K\\]"
		w.saveCursorPosition = "\\[\x1b7\\]"
		w.restoreCursorPosition = "\\[\x1b8\\]"
		w.title = "\\[\x1b]0;%s\007\\]"
		w.escapeLeft = "\\["
		w.escapeRight = "\\]"
		w.hyperlink = "\\[\x1b]8;;%s\x1b\\\\\\]%s\\[\x1b]8;;\x1b\\\\\\]"
		w.hyperlinkRegex = `(?P<STR>\\\[\x1b\]8;;(.+)\x1b\\\\\\\](?P<TEXT>.+)\\\[\x1b\]8;;\x1b\\\\\\\])`
		w.osc99 = "\\[\x1b]9;9;\"%s\"\x1b\\\\\\]"
		w.osc7 = "\\[\x1b]7;\"file://%s/%s\"\x1b\\\\\\]"
		w.osc51 = "\\[\x1b]51;A;%s@%s:%s\x1b\\\\\\]"
	case "zsh":
		w.format = "%%{%s%%}"
		w.linechange = "%%{\x1b[%d%s%%}"
		w.right = "%%{\x1b[%dC%%}"
		w.left = "%%{\x1b[%dD%%}"
		w.clearBelow = "%{\x1b[0J%}"
		w.clearLine = "%{\x1b[K%}"
		w.saveCursorPosition = "%{\x1b7%}"
		w.restoreCursorPosition = "%{\x1b8%}"
		w.title = "%%{\x1b]0;%s\007%%}"
		w.escapeLeft = "%{"
		w.escapeRight = "%}"
		w.hyperlink = "%%{\x1b]8;;%s\x1b\\%%}%s%%{\x1b]8;;\x1b\\%%}"
		w.hyperlinkRegex = `(?P<STR>%{\x1b]8;;(.+)\x1b\\%}(?P<TEXT>.+)%{\x1b]8;;\x1b\\%})`
		w.osc99 = "%%{\x1b]9;9;\"%s\"\x1b\\%%}"
		w.osc7 = "%%{\x1b]7;file:\"//%s/%s\"\x1b\\%%}"
		w.osc51 = "%%{\x1b]51;A%s@%s:%s\x1b\\%%}"
	default:
		w.linechange = "\x1b[%d%s"
		w.right = "\x1b[%dC"
		w.left = "\x1b[%dD"
		w.clearBelow = "\x1b[0J"
		w.clearLine = "\x1b[K"
		w.saveCursorPosition = "\x1b7"
		w.restoreCursorPosition = "\x1b8"
		w.title = "\x1b]0;%s\007"
		// when in fish on Linux, it seems hyperlinks ending with \\ print a \
		// unlike on macOS. However, this is a fish bug, so do not try to fix it here:
		// https://github.com/JanDeDobbeleer/oh-my-posh/pull/3288#issuecomment-1369137068
		w.hyperlink = "\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\"
		w.hyperlinkRegex = "(?P<STR>\x1b]8;;(.+)\x1b\\\\\\\\?(?P<TEXT>.+)\x1b]8;;\x1b\\\\)"
		w.osc99 = "\x1b]9;9;\"%s\"\x1b\\"
		w.osc7 = "\x1b]7;\"file://%s/%s\"\x1b\\"
		w.osc51 = "\x1b]51;A%s@%s:%s\x1b\\"
	}
}

func (w *Writer) SetColors(background, foreground string) {
	w.Colors = &cachedColor{
		Background: background,
		Foreground: foreground,
	}
}

func (w *Writer) SetParentColors(background, foreground string) {
	if w.ParentColors == nil {
		w.ParentColors = make([]*cachedColor, 0)
	}
	w.ParentColors = append([]*cachedColor{{
		Background: background,
		Foreground: foreground,
	}}, w.ParentColors...)
}

func (w *Writer) CarriageForward() string {
	return fmt.Sprintf(w.right, 1000)
}

func (w *Writer) GetCursorForRightWrite(length, offset int) string {
	strippedLen := length + (-offset)
	return fmt.Sprintf(w.left, strippedLen)
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
	return fmt.Sprintf(w.linechange, numberOfLines, position)
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
		return fmt.Sprintf(w.osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(w.osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(w.osc99, pwd)
	}
}

func (w *Writer) ClearAfter() string {
	if w.Plain {
		return ""
	}
	return w.clearLine + w.clearBelow
}

func (w *Writer) FormatTitle(title string) string {
	title = w.trimAnsi(title)
	// we have to do this to prevent bash/zsh from misidentifying escape sequences
	switch w.shell {
	case shell.BASH:
		title = strings.NewReplacer("`", "\\`", `\`, `\\`).Replace(title)
	case shell.ZSH:
		title = strings.NewReplacer("`", "\\`", `%`, `%%`).Replace(title)
	}
	return fmt.Sprintf(w.title, title)
}

func (w *Writer) FormatText(text string) string {
	return fmt.Sprintf(w.format, text)
}

func (w *Writer) SaveCursorPosition() string {
	return w.saveCursorPosition
}

func (w *Writer) RestoreCursorPosition() string {
	return w.restoreCursorPosition
}

func (w *Writer) LineBreak() string {
	cr := fmt.Sprintf(w.left, 1000)
	lf := fmt.Sprintf(w.linechange, 1, "B")
	return cr + lf
}

func (w *Writer) Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	if !w.Plain {
		text = w.GenerateHyperlink(text)
	}

	w.background, w.foreground = w.asAnsiColors(background, foreground)
	// default to white foreground
	if w.foreground.IsEmpty() {
		w.foreground = w.AnsiColors.AnsiColorFromString("white", false)
	}
	// validate if we start with a color override
	match := regex.FindNamedRegexMatch(anchorRegex, text)
	if len(match) != 0 {
		colorOverride := true
		for _, style := range knownStyles {
			if match["ANCHOR"] != style.AnchorStart {
				continue
			}
			w.printEscapedAnsiString(style.Start)
			colorOverride = false
		}
		if colorOverride {
			w.currentBackground, w.currentForeground = w.asAnsiColors(match["BG"], match["FG"])
		}
	}
	w.writeSegmentColors()

	text = text[len(match["ANCHOR"]):]
	w.runes = []rune(text)

	for i := 0; i < len(w.runes); i++ {
		s := w.runes[i]
		// ignore everything which isn't overriding
		if s != '<' {
			w.length += runewidth.RuneWidth(s)
			w.builder.WriteRune(s)
			continue
		}

		// color/end overrides first
		text = string(w.runes[i:])
		match = regex.FindNamedRegexMatch(anchorRegex, text)
		if len(match) > 0 {
			i = w.writeColorOverrides(match, background, i)
			continue
		}

		w.length += runewidth.RuneWidth(s)
		w.builder.WriteRune(s)
	}

	w.printEscapedAnsiString(colorStyle.End)

	// reset current
	w.currentBackground = ""
	w.currentForeground = ""
}

func (w *Writer) printEscapedAnsiString(text string) {
	if w.Plain {
		return
	}
	if len(w.format) == 0 {
		w.builder.WriteString(text)
		return
	}
	w.builder.WriteString(fmt.Sprintf(w.format, text))
}

func (w *Writer) getAnsiFromColorString(colorString string, isBackground bool) Color {
	return w.AnsiColors.AnsiColorFromString(colorString, isBackground)
}

func (w *Writer) writeSegmentColors() {
	// use correct starting colors
	bg := w.background
	fg := w.foreground
	if !w.currentBackground.IsEmpty() {
		bg = w.currentBackground
	}
	if !w.currentForeground.IsEmpty() {
		fg = w.currentForeground
	}

	if fg.IsTransparent() && len(w.TerminalBackground) != 0 {
		background := w.getAnsiFromColorString(w.TerminalBackground, false)
		w.printEscapedAnsiString(fmt.Sprintf(colorise, background))
		w.printEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	} else if fg.IsTransparent() && !bg.IsEmpty() {
		w.printEscapedAnsiString(fmt.Sprintf(transparent, bg))
	} else {
		if !bg.IsEmpty() && !bg.IsTransparent() {
			w.printEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}
		if !fg.IsEmpty() {
			w.printEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	w.currentBackground = bg
	w.currentForeground = fg
}

func (w *Writer) writeColorOverrides(match map[string]string, background string, i int) (position int) {
	position = i
	// check color reset first
	if match["ANCHOR"] == colorStyle.AnchorEnd {
		// make sure to reset the colors if needed
		position += len([]rune(colorStyle.AnchorEnd)) - 1
		// do not restore colors at the end of the string, we print it anyways
		if position == len(w.runes)-1 {
			return
		}
		if w.currentBackground != w.background {
			w.printEscapedAnsiString(fmt.Sprintf(colorise, w.background))
		}
		if w.currentForeground != w.foreground {
			w.printEscapedAnsiString(fmt.Sprintf(colorise, w.foreground))
		}
		return
	}

	position += len([]rune(match["ANCHOR"])) - 1

	for _, style := range knownStyles {
		if style.AnchorEnd == match["ANCHOR"] {
			w.printEscapedAnsiString(style.End)
			return
		}
		if style.AnchorStart == match["ANCHOR"] {
			w.printEscapedAnsiString(style.Start)
			return
		}
	}

	if match["FG"] == Transparent && len(match["BG"]) == 0 {
		match["BG"] = background
	}
	w.currentBackground, w.currentForeground = w.asAnsiColors(match["BG"], match["FG"])

	// make sure we have colors
	if w.currentForeground.IsEmpty() {
		w.currentForeground = w.foreground
	}
	if w.currentBackground.IsEmpty() {
		w.currentBackground = w.background
	}

	if w.currentForeground.IsTransparent() && len(w.TerminalBackground) != 0 {
		background := w.getAnsiFromColorString(w.TerminalBackground, false)
		w.printEscapedAnsiString(fmt.Sprintf(colorise, background))
		w.printEscapedAnsiString(fmt.Sprintf(colorise, w.currentBackground.ToForeground()))
		return
	}

	if w.currentForeground.IsTransparent() && !w.currentBackground.IsTransparent() {
		w.printEscapedAnsiString(fmt.Sprintf(transparent, w.currentBackground))
		return
	}

	if w.currentBackground != w.background {
		// end the colors in case we have a transparent background
		if w.currentBackground.IsTransparent() {
			w.printEscapedAnsiString(colorStyle.End)
		} else {
			w.printEscapedAnsiString(fmt.Sprintf(colorise, w.currentBackground))
		}
	}

	if w.currentForeground != w.foreground || w.currentBackground.IsTransparent() {
		w.printEscapedAnsiString(fmt.Sprintf(colorise, w.currentForeground))
	}

	return position
}

func (w *Writer) asAnsiColors(background, foreground string) (Color, Color) {
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
		case keyword == Background && w.Colors != nil:
			return w.Colors.Background
		case keyword == Foreground && w.Colors != nil:
			return w.Colors.Foreground
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

func (w *Writer) String() (string, int) {
	defer func() {
		w.length = 0
		w.builder.Reset()
	}()

	return w.builder.String(), w.length
}
