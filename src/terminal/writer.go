package terminal

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
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

	BackgroundColor string
	CurrentColors   *Colors
	ParentColors    []*Colors
	AnsiColors      ColorString

	Plain       bool
	Interactive bool

	trueColor bool

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

	Shell   string
	Program string

	formats *shellFormats
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

	WindowsTerminal = "Windows Terminal"
	Warp            = "WarpTerminal"
	ITerm           = "iTerm.app"
	AppleTerminal   = "Apple_Terminal"
	Unknown         = "Unknown"
)

func Init(sh string) {
	Shell = sh
	Program = getTerminalName()

	log.Debug("Terminal shell: %s", Shell)
	log.Debug("Terminal program: %s", Program)

	trueColor = Program != AppleTerminal

	switch Shell {
	case shell.BASH:
		formats = &shellFormats{
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
		formats = &shellFormats{
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
		formats = &shellFormats{
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

	if Shell == shell.ZSH {
		formats.escapeSequences = map[rune]rune{
			96: 92, // backtick
			37: 37, // %
		}
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

func SetColors(background, foreground string) {
	CurrentColors = &Colors{
		Background: background,
		Foreground: foreground,
	}
}

func SetParentColors(background, foreground string) {
	if ParentColors == nil {
		ParentColors = make([]*Colors, 0)
	}

	ParentColors = append([]*Colors{{
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

	return fmt.Sprintf(formats.linechange, numberOfLines, position)
}

func Pwd(pwdType, userName, hostName, pwd string) string {
	if Plain {
		return ""
	}

	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}

	switch pwdType {
	case OSC7:
		return fmt.Sprintf(formats.osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(formats.osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(formats.osc99, pwd)
	}
}

func ClearAfter() string {
	if Plain {
		return ""
	}

	return formats.clearLine + formats.clearBelow
}

func FormatTitle(title string) string {
	title = trimAnsi(title)

	if Plain {
		return title
	}

	// we have to do this to prevent bash/zsh from misidentifying escape sequences
	switch Shell {
	case shell.BASH:
		title = strings.NewReplacer("`", "\\`", `\`, `\\`).Replace(title)
	case shell.ZSH:
		title = strings.NewReplacer("`", "\\`", `%`, `%%`).Replace(title)
	case shell.ELVISH, shell.XONSH:
		// these shells don't support setting the title
		return ""
	}

	return fmt.Sprintf(formats.title, title)
}

func EscapeText(text string) string {
	return fmt.Sprintf(formats.escape, text)
}

func SaveCursorPosition() string {
	return formats.saveCursorPosition
}

func RestoreCursorPosition() string {
	return formats.restoreCursorPosition
}

func PromptStart() string {
	return fmt.Sprintf(formats.escape, "\x1b]133;A\007")
}

func CommandStart() string {
	return fmt.Sprintf(formats.escape, "\x1b]133;B\007")
}

func CommandFinished(code int, ignore bool) string {
	if ignore {
		return fmt.Sprintf(formats.escape, "\x1b]133;D\007")
	}

	mark := fmt.Sprintf("\x1b]133;D;%d\007", code)

	return fmt.Sprintf(formats.escape, mark)
}

func LineBreak() string {
	cr := fmt.Sprintf(formats.left, 1000)
	lf := fmt.Sprintf(formats.linechange, 1, "B")
	return cr + lf
}

func Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	backgroundColor, foregroundColor = asAnsiColors(background, foreground)

	// default to white foreground
	if foregroundColor.IsEmpty() {
		foregroundColor = AnsiColors.ToColor("white", false)
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
			currentColor.Add(asAnsiColors(match[BG], match[FG]))
		}
	}

	writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match[ANCHOR] == hyperLinkStart {
		isHyperlink = true
		builder.WriteString(formats.hyperlinkStart)
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
				builder.WriteString(formats.hyperlinkStart)
				continue
			case hyperLinkText:
				isHyperlink = false
				i += len([]rune(match[ANCHOR])) - 1
				hyperlinkTextPosition = i
				builder.WriteString(formats.hyperlinkCenter)
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
				builder.WriteString(formats.hyperlinkEnd)
				continue
			}

			i = writeArchorOverride(match, background, i)
			continue
		}

		length += runewidth.RuneWidth(s)
		write(s)
	}

	// reset colors
	writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	currentColor.Pop()
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

	if len(formats.escape) != 0 {
		text = fmt.Sprintf(formats.escape, text)
	}

	builder.WriteString(text)
}

func getAnsiFromColorString(colorString string, isBackground bool) Color {
	return AnsiColors.ToColor(colorString, isBackground)
}

func write(s rune) {
	if isInvisible {
		return
	}

	if isHyperlink {
		builder.WriteRune(s)
		return
	}

	if !Interactive {
		for special, escape := range formats.escapeSequences {
			if s == special && lastRune != escape {
				builder.WriteRune(escape)
			}
		}
	}

	length += runewidth.RuneWidth(s)
	lastRune = s
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

	if fg.IsTransparent() && len(BackgroundColor) != 0 { //nolint: gocritic
		background := getAnsiFromColorString(BackgroundColor, false)
		writeEscapedAnsiString(fmt.Sprintf(colorise, background))
		writeEscapedAnsiString(fmt.Sprintf(colorise, bg.ToForeground()))
	} else if fg.IsTransparent() && !bg.IsEmpty() {
		isTransparent = true
		writeEscapedAnsiString(fmt.Sprintf(transparentStart, bg))
	} else {
		if !bg.IsEmpty() && !bg.IsTransparent() {
			writeEscapedAnsiString(fmt.Sprintf(colorise, bg))
		}
		if !fg.IsEmpty() {
			writeEscapedAnsiString(fmt.Sprintf(colorise, fg))
		}
	}

	// set current colors
	currentColor.Add(bg, fg)
}

func writeArchorOverride(match map[string]string, background string, i int) int {
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

	if match[FG] == Transparent && len(match[BG]) == 0 {
		match[BG] = background
	}

	bg, fg := asAnsiColors(match[BG], match[FG])

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
		background := getAnsiFromColorString(BackgroundColor, false)
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

func asAnsiColors(background, foreground string) (Color, Color) {
	if len(background) == 0 {
		background = Background
	}

	if len(foreground) == 0 {
		foreground = Foreground
	}

	background = expandKeyword(background)
	foreground = expandKeyword(foreground)

	inverted := foreground == Transparent && len(background) != 0

	backgroundAnsi := getAnsiFromColorString(background, !inverted)
	foregroundAnsi := getAnsiFromColorString(foreground, false)

	return backgroundAnsi, foregroundAnsi
}

func isKeyword(color string) bool {
	switch color {
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func expandKeyword(keyword string) string {
	resolveParentColor := func(keyword string) string {
		for _, color := range ParentColors {
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
		case keyword == Background && CurrentColors != nil:
			return CurrentColors.Background
		case keyword == Foreground && CurrentColors != nil:
			return CurrentColors.Foreground
		case (keyword == ParentBackground || keyword == ParentForeground) && ParentColors != nil:
			return resolveParentColor(keyword)
		default:
			return Transparent
		}
	}

	for ok := isKeyword(keyword); ok; ok = isKeyword(keyword) {
		resolved := resolveKeyword(keyword)
		if resolved == keyword {
			break
		}

		keyword = resolved
	}

	return keyword
}

func trimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}
