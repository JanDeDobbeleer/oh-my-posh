package color

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/regex"
	"github.com/jandedobbeleer/oh-my-posh/shell"
	"github.com/mattn/go-runewidth"
)

type Writer interface {
	Init(shellName string)
	Write(background, foreground, text string)
	String() (string, int)
	SetColors(background, foreground string)
	SetParentColors(background, foreground string)
	CarriageForward() string
	GetCursorForRightWrite(length, offset int) string
	ChangeLine(numberOfLines int) string
	ConsolePwd(pwdType, userName, hostName, pwd string) string
	ClearAfter() string
	FormatTitle(title string) string
	FormatText(text string) string
	SaveCursorPosition() string
	RestoreCursorPosition() string
	LineBreak() string
	TrimAnsi(text string) string
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
	colorStyle = &style{AnchorStart: "COLOR", AnchorEnd: `</>`, End: "\x1b[0m"}
)

type style struct {
	AnchorStart string
	AnchorEnd   string
	Start       string
	End         string
}

type Color struct {
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

// AnsiWriter writes colorized ANSI strings
type AnsiWriter struct {
	TerminalBackground string
	Colors             *Color
	ParentColors       []*Color
	AnsiColors         AnsiColors
	Plain              bool

	builder strings.Builder
	length  int

	foreground        AnsiColor
	background        AnsiColor
	currentForeground AnsiColor
	currentBackground AnsiColor
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

func (a *AnsiWriter) Init(shellName string) {
	a.shell = shellName
	switch a.shell {
	case shell.BASH:
		a.format = "\\[%s\\]"
		a.linechange = "\\[\x1b[%d%s\\]"
		a.right = "\\[\x1b[%dC\\]"
		a.left = "\\[\x1b[%dD\\]"
		a.clearBelow = "\\[\x1b[0J\\]"
		a.clearLine = "\\[\x1b[K\\]"
		a.saveCursorPosition = "\\[\x1b7\\]"
		a.restoreCursorPosition = "\\[\x1b8\\]"
		a.title = "\\[\x1b]0;%s\007\\]"
		a.escapeLeft = "\\["
		a.escapeRight = "\\]"
		a.hyperlink = "\\[\x1b]8;;%s\x1b\\\\\\]%s\\[\x1b]8;;\x1b\\\\\\]"
		a.hyperlinkRegex = `(?P<STR>\\\[\x1b\]8;;(.+)\x1b\\\\\\\](?P<TEXT>.+)\\\[\x1b\]8;;\x1b\\\\\\\])`
		a.osc99 = "\\[\x1b]9;9;\"%s\"\x1b\\\\\\]"
		a.osc7 = "\\[\x1b]7;\"file://%s/%s\"\x1b\\\\\\]"
		a.osc51 = "\\[\x1b]51;A;%s@%s:%s\x1b\\\\\\]"
	case "zsh":
		a.format = "%%{%s%%}"
		a.linechange = "%%{\x1b[%d%s%%}"
		a.right = "%%{\x1b[%dC%%}"
		a.left = "%%{\x1b[%dD%%}"
		a.clearBelow = "%{\x1b[0J%}"
		a.clearLine = "%{\x1b[K%}"
		a.saveCursorPosition = "%{\x1b7%}"
		a.restoreCursorPosition = "%{\x1b8%}"
		a.title = "%%{\x1b]0;%s\007%%}"
		a.escapeLeft = "%{"
		a.escapeRight = "%}"
		a.hyperlink = "%%{\x1b]8;;%s\x1b\\%%}%s%%{\x1b]8;;\x1b\\%%}"
		a.hyperlinkRegex = `(?P<STR>%{\x1b]8;;(.+)\x1b\\%}(?P<TEXT>.+)%{\x1b]8;;\x1b\\%})`
		a.osc99 = "%%{\x1b]9;9;\"%s\"\x1b\\%%}"
		a.osc7 = "%%{\x1b]7;file:\"//%s/%s\"\x1b\\%%}"
		a.osc51 = "%%{\x1b]51;A%s@%s:%s\x1b\\%%}"
	default:
		a.linechange = "\x1b[%d%s"
		a.right = "\x1b[%dC"
		a.left = "\x1b[%dD"
		a.clearBelow = "\x1b[0J"
		a.clearLine = "\x1b[K"
		a.saveCursorPosition = "\x1b7"
		a.restoreCursorPosition = "\x1b8"
		a.title = "\x1b]0;%s\007"
		// when in fish on Linux, it seems hyperlinks ending with \\ print a \
		// unlike on macOS. However, this is a fish bug, so do not try to fix it here:
		// https://github.com/JanDeDobbeleer/oh-my-posh/pull/3288#issuecomment-1369137068
		a.hyperlink = "\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\"
		a.hyperlinkRegex = "(?P<STR>\x1b]8;;(.+)\x1b\\\\\\\\?(?P<TEXT>.+)\x1b]8;;\x1b\\\\)"
		a.osc99 = "\x1b]9;9;\"%s\"\x1b\\"
		a.osc7 = "\x1b]7;\"file://%s/%s\"\x1b\\"
		a.osc51 = "\x1b]51;A%s@%s:%s\x1b\\"
	}
}

func (a *AnsiWriter) SetColors(background, foreground string) {
	a.Colors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

func (a *AnsiWriter) SetParentColors(background, foreground string) {
	if a.ParentColors == nil {
		a.ParentColors = make([]*Color, 0)
	}
	a.ParentColors = append([]*Color{{
		Background: background,
		Foreground: foreground,
	}}, a.ParentColors...)
}

func (a *AnsiWriter) CarriageForward() string {
	return fmt.Sprintf(a.right, 1000)
}

func (a *AnsiWriter) GetCursorForRightWrite(length, offset int) string {
	strippedLen := length + (-offset)
	return fmt.Sprintf(a.left, strippedLen)
}

func (a *AnsiWriter) ChangeLine(numberOfLines int) string {
	if a.Plain {
		return ""
	}
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(a.linechange, numberOfLines, position)
}

func (a *AnsiWriter) ConsolePwd(pwdType, userName, hostName, pwd string) string {
	if a.Plain {
		return ""
	}
	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}
	switch pwdType {
	case OSC7:
		return fmt.Sprintf(a.osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(a.osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(a.osc99, pwd)
	}
}

func (a *AnsiWriter) ClearAfter() string {
	if a.Plain {
		return ""
	}
	return a.clearLine + a.clearBelow
}

func (a *AnsiWriter) FormatTitle(title string) string {
	title = a.TrimAnsi(title)
	// we have to do this to prevent bash/zsh from misidentifying escape sequences
	switch a.shell {
	case shell.BASH:
		title = strings.NewReplacer("`", "\\`", `\`, `\\`).Replace(title)
	case shell.ZSH:
		title = strings.NewReplacer("`", "\\`", `%`, `%%`).Replace(title)
	}
	return fmt.Sprintf(a.title, title)
}

func (a *AnsiWriter) FormatText(text string) string {
	return fmt.Sprintf(a.format, text)
}

func (a *AnsiWriter) SaveCursorPosition() string {
	return a.saveCursorPosition
}

func (a *AnsiWriter) RestoreCursorPosition() string {
	return a.restoreCursorPosition
}

func (a *AnsiWriter) LineBreak() string {
	cr := fmt.Sprintf(a.left, 1000)
	lf := fmt.Sprintf(a.linechange, 1, "B")
	return cr + lf
}

func (a *AnsiWriter) Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	if !a.Plain {
		text = a.GenerateHyperlink(text)
	}

	a.background, a.foreground = a.asAnsiColors(background, foreground)
	// default to white foreground
	if a.foreground.IsEmpty() {
		a.foreground = a.AnsiColors.AnsiColorFromString("white", false)
	}
	// validate if we start with a color override
	match := regex.FindNamedRegexMatch(anchorRegex, text)
	if len(match) != 0 {
		colorOverride := true
		for _, style := range knownStyles {
			if match["ANCHOR"] != style.AnchorStart {
				continue
			}
			a.printEscapedAnsiString(style.Start)
			colorOverride = false
		}
		if colorOverride {
			a.currentBackground, a.currentForeground = a.asAnsiColors(match["BG"], match["FG"])
		}
	}
	a.writeSegmentColors()

	text = text[len(match["ANCHOR"]):]
	a.runes = []rune(text)

	for i := 0; i < len(a.runes); i++ {
		s := a.runes[i]
		// ignore everything which isn't overriding
		if s != '<' {
			a.length += runewidth.RuneWidth(s)
			a.builder.WriteRune(s)
			continue
		}

		// color/end overrides first
		text = string(a.runes[i:])
		match = regex.FindNamedRegexMatch(anchorRegex, text)
		if len(match) > 0 {
			i = a.writeColorOverrides(match, background, i)
			continue
		}

		a.length += runewidth.RuneWidth(s)
		a.builder.WriteRune(s)
	}

	a.printEscapedAnsiString(colorStyle.End)

	// reset current
	a.currentBackground = ""
	a.currentForeground = ""
}

func (a *AnsiWriter) printEscapedAnsiString(text string) {
	if a.Plain {
		return
	}
	if len(a.format) == 0 {
		a.builder.WriteString(text)
		return
	}
	a.builder.WriteString(fmt.Sprintf(a.format, text))
}

func (a *AnsiWriter) getAnsiFromColorString(colorString string, isBackground bool) AnsiColor {
	return a.AnsiColors.AnsiColorFromString(colorString, isBackground)
}

func (a *AnsiWriter) writeSegmentColors() {
	// use correct starting colors
	bg := a.background
	fg := a.foreground
	if !a.currentBackground.IsEmpty() {
		bg = a.currentBackground
	}
	if !a.currentForeground.IsEmpty() {
		fg = a.currentForeground
	}

	if fg.IsTransparent() && len(a.TerminalBackground) != 0 {
		background := a.getAnsiFromColorString(a.TerminalBackground, false)
		a.printEscapedAnsiString(fmt.Sprintf(colorise, background))
		a.printEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	} else if fg.IsTransparent() && !bg.IsEmpty() {
		a.printEscapedAnsiString(fmt.Sprintf(transparent, bg))
	} else {
		if !bg.IsEmpty() && !bg.IsTransparent() {
			a.printEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}
		if !fg.IsEmpty() {
			a.printEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	a.currentBackground = bg
	a.currentForeground = fg
}

func (a *AnsiWriter) writeColorOverrides(match map[string]string, background string, i int) (position int) {
	position = i
	// check color reset first
	if match["ANCHOR"] == colorStyle.AnchorEnd {
		// make sure to reset the colors if needed
		position += len([]rune(colorStyle.AnchorEnd)) - 1
		// do not restore colors at the end of the string, we print it anyways
		if position == len(a.runes)-1 {
			return
		}
		if a.currentBackground != a.background {
			a.printEscapedAnsiString(fmt.Sprintf(colorise, a.background))
		}
		if a.currentForeground != a.foreground {
			a.printEscapedAnsiString(fmt.Sprintf(colorise, a.foreground))
		}
		return
	}

	position += len([]rune(match["ANCHOR"])) - 1

	for _, style := range knownStyles {
		if style.AnchorEnd == match["ANCHOR"] {
			a.printEscapedAnsiString(style.End)
			return
		}
		if style.AnchorStart == match["ANCHOR"] {
			a.printEscapedAnsiString(style.Start)
			return
		}
	}

	if match["FG"] == Transparent && len(match["BG"]) == 0 {
		match["BG"] = background
	}
	a.currentBackground, a.currentForeground = a.asAnsiColors(match["BG"], match["FG"])

	// make sure we have colors
	if a.currentForeground.IsEmpty() {
		a.currentForeground = a.foreground
	}
	if a.currentBackground.IsEmpty() {
		a.currentBackground = a.background
	}

	if a.currentForeground.IsTransparent() && len(a.TerminalBackground) != 0 {
		background := a.getAnsiFromColorString(a.TerminalBackground, false)
		a.printEscapedAnsiString(fmt.Sprintf(colorise, background))
		a.printEscapedAnsiString(fmt.Sprintf(colorise, a.currentBackground.ToForeground()))
		return
	}

	if a.currentForeground.IsTransparent() && !a.currentBackground.IsTransparent() {
		a.printEscapedAnsiString(fmt.Sprintf(transparent, a.currentBackground))
		return
	}

	if a.currentBackground != a.background {
		// end the colors in case we have a transparent background
		if a.currentBackground.IsTransparent() {
			a.printEscapedAnsiString(colorStyle.End)
		} else {
			a.printEscapedAnsiString(fmt.Sprintf(colorise, a.currentBackground))
		}
	}

	if a.currentForeground != a.foreground || a.currentBackground.IsTransparent() {
		a.printEscapedAnsiString(fmt.Sprintf(colorise, a.currentForeground))
	}

	return position
}

func (a *AnsiWriter) asAnsiColors(background, foreground string) (AnsiColor, AnsiColor) {
	background = a.expandKeyword(background)
	foreground = a.expandKeyword(foreground)
	inverted := foreground == Transparent && len(background) != 0
	backgroundAnsi := a.getAnsiFromColorString(background, !inverted)
	foregroundAnsi := a.getAnsiFromColorString(foreground, false)
	return backgroundAnsi, foregroundAnsi
}

func (a *AnsiWriter) isKeyword(color string) bool {
	switch color {
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func (a *AnsiWriter) expandKeyword(keyword string) string {
	resolveParentColor := func(keyword string) string {
		for _, color := range a.ParentColors {
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
		case keyword == Background && a.Colors != nil:
			return a.Colors.Background
		case keyword == Foreground && a.Colors != nil:
			return a.Colors.Foreground
		case (keyword == ParentBackground || keyword == ParentForeground) && a.ParentColors != nil:
			return resolveParentColor(keyword)
		default:
			return Transparent
		}
	}
	for ok := a.isKeyword(keyword); ok; ok = a.isKeyword(keyword) {
		resolved := resolveKeyword(keyword)
		if resolved == keyword {
			break
		}
		keyword = resolved
	}
	return keyword
}

func (a *AnsiWriter) String() (string, int) {
	defer func() {
		a.length = 0
		a.builder.Reset()
	}()

	return a.builder.String(), a.length
}
