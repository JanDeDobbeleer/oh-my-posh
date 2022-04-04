package color

import (
	"fmt"
	"oh-my-posh/regex"
	"oh-my-posh/shell"
	"strings"
)

const (
	AnsiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

type Ansi struct {
	title                 string
	shell                 string
	linechange            string
	left                  string
	right                 string
	creset                string
	clearBelow            string
	clearLine             string
	saveCursorPosition    string
	restoreCursorPosition string
	colorSingle           string
	colorFull             string
	colorTransparent      string
	escapeLeft            string
	escapeRight           string
	hyperlink             string
	hyperlinkRegex        string
	osc99                 string
	bold                  string
	italic                string
	underline             string
	strikethrough         string
	blink                 string
	reverse               string
	dimmed                string
	format                string
	reservedSequences     []sequenceReplacement
}

type sequenceReplacement struct {
	text        string
	replacement string
}

func (a *Ansi) Init(shellName string) {
	a.shell = shellName
	a.initEscapeSequences(shellName)
	switch shellName {
	case shell.ZSH:
		a.format = "%%{%s%%}"
		a.linechange = "%%{\x1b[%d%s%%}"
		a.right = "%%{\x1b[%dC%%}"
		a.left = "%%{\x1b[%dD%%}"
		a.creset = "%{\x1b[0m%}"
		a.clearBelow = "%{\x1b[0J%}"
		a.clearLine = "%{\x1b[K%}"
		a.saveCursorPosition = "%{\x1b7%}"
		a.restoreCursorPosition = "%{\x1b8%}"
		a.title = "%%{\x1b]0;%s\007%%}"
		a.colorSingle = "%%{\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorFull = "%%{\x1b[%sm\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorTransparent = "%%{\x1b[%s;49m\x1b[7m%%}%s%%{\x1b[0m%%}"
		a.escapeLeft = "%{"
		a.escapeRight = "%}"
		a.hyperlink = "%%{\x1b]8;;%s\x1b\\\\%%}%s%%{\x1b]8;;\x1b\\\\%%}"
		a.hyperlinkRegex = `(?P<STR>%{\x1b]8;;(.+)\x1b\\\\%}(?P<TEXT>.+)%{\x1b]8;;\x1b\\\\%})`
		a.osc99 = "%%{\x1b]9;9;\"%s\"\x1b\\%%}"
		a.bold = "%%{\x1b[1m%%}%s%%{\x1b[22m%%}"
		a.italic = "%%{\x1b[3m%%}%s%%{\x1b[23m%%}"
		a.underline = "%%{\x1b[4m%%}%s%%{\x1b[24m%%}"
		a.blink = "%%{\x1b[5m%%}%s%%{\x1b[25m%%}"
		a.reverse = "%%{\x1b[7m%%}%s%%{\x1b[27m%%}"
		a.dimmed = "%%{\x1b[2m%%}%s%%{\x1b[22m%%}"
		a.strikethrough = "%%{\x1b[9m%%}%s%%{\x1b[29m%%}"
	case shell.BASH:
		a.format = "\\[%s\\]"
		a.linechange = "\\[\x1b[%d%s\\]"
		a.right = "\\[\x1b[%dC\\]"
		a.left = "\\[\x1b[%dD\\]"
		a.creset = "\\[\x1b[0m\\]"
		a.clearBelow = "\\[\x1b[0J\\]"
		a.clearLine = "\\[\x1b[K\\]"
		a.saveCursorPosition = "\\[\x1b7\\]"
		a.restoreCursorPosition = "\\[\x1b8\\]"
		a.title = "\\[\x1b]0;%s\007\\]"
		a.colorSingle = "\\[\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorFull = "\\[\x1b[%sm\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorTransparent = "\\[\x1b[%s;49m\x1b[7m\\]%s\\[\x1b[0m\\]"
		a.escapeLeft = "\\["
		a.escapeRight = "\\]"
		a.hyperlink = "\\[\x1b]8;;%s\x1b\\\\\\]%s\\[\x1b]8;;\x1b\\\\\\]"
		a.hyperlinkRegex = `(?P<STR>\\\[\x1b\]8;;(.+)\x1b\\\\\\\](?P<TEXT>.+)\\\[\x1b\]8;;\x1b\\\\\\\])`
		a.osc99 = "\\[\x1b]9;9;\"%s\"\x1b\\\\\\]"
		a.bold = "\\[\x1b[1m\\]%s\\[\x1b[22m\\]"
		a.italic = "\\[\x1b[3m\\]%s\\[\x1b[23m\\]"
		a.underline = "\\[\x1b[4m\\]%s\\[\x1b[24m\\]"
		a.blink = "\\[\x1b[5m%s\\[\x1b[25m\\]"
		a.reverse = "\\[\x1b[7m\\]%s\\[\x1b[27m\\]"
		a.dimmed = "\\[\x1b[2m\\]%s\\[\x1b[22m\\]"
		a.strikethrough = "\\[\x1b[9m\\]%s\\[\x1b[29m\\]"
	default:
		a.format = "%s"
		a.linechange = "\x1b[%d%s"
		a.right = "\x1b[%dC"
		a.left = "\x1b[%dD"
		a.creset = "\x1b[0m"
		a.clearBelow = "\x1b[0J"
		a.clearLine = "\x1b[K"
		a.saveCursorPosition = "\x1b7"
		a.restoreCursorPosition = "\x1b8"
		a.title = "\x1b]0;%s\007"
		a.colorSingle = "\x1b[%sm%s\x1b[0m"
		a.colorFull = "\x1b[%sm\x1b[%sm%s\x1b[0m"
		a.colorTransparent = "\x1b[%s;49m\x1b[7m%s\x1b[0m"
		a.escapeLeft = ""
		a.escapeRight = ""
		a.hyperlink = "\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\"
		a.hyperlinkRegex = "(?P<STR>\x1b]8;;(.+)\x1b\\\\\\\\?(?P<TEXT>.+)\x1b]8;;\x1b\\\\)"
		a.osc99 = "\x1b]9;9;\"%s\"\x1b\\"
		a.bold = "\x1b[1m%s\x1b[22m"
		a.italic = "\x1b[3m%s\x1b[23m"
		a.underline = "\x1b[4m%s\x1b[24m"
		a.blink = "\x1b[5m%s\x1b[25m"
		a.reverse = "\x1b[7m%s\x1b[27m"
		a.dimmed = "\x1b[2m%s\x1b[22m"
		a.strikethrough = "\x1b[9m%s\x1b[29m"
	}
}

func (a *Ansi) InitPlain(shellName string) {
	a.Init(shell.PLAIN)
	a.initEscapeSequences(shellName)
}

func (a *Ansi) initEscapeSequences(shellName string) {
	switch shellName {
	case shell.ZSH:
		// escape double quotes and variable expansion
		a.reservedSequences = []sequenceReplacement{
			{text: "`", replacement: "'"},
			{text: `%l`, replacement: `%%l`},
			{text: `%M`, replacement: `%%M`},
			{text: `%m`, replacement: `%%m`},
			{text: `%n`, replacement: `%%n`},
			{text: `%y`, replacement: `%%y`},
			{text: `%#`, replacement: `%%#`},
			{text: `%?`, replacement: `%%?`},
			{text: `%_`, replacement: `%%_`},
			{text: `%^`, replacement: `%%^`},
			{text: `%d`, replacement: `%%d`},
			{text: `%/`, replacement: `%%/`},
			{text: `%~`, replacement: `%%~`},
			{text: `%e`, replacement: `%%e`},
			{text: `%h`, replacement: `%%h`},
			{text: `%!`, replacement: `%%!`},
			{text: `%i`, replacement: `%%i`},
			{text: `%I`, replacement: `%%I`},
			{text: `%j`, replacement: `%%j`},
			{text: `%L`, replacement: `%%L`},
			{text: `%N`, replacement: `%%N`},
			{text: `%x`, replacement: `%%x`},
			{text: `%c`, replacement: `%%c`},
			{text: `%.`, replacement: `%%.`},
			{text: `%C`, replacement: `%%C`},
			{text: `%D`, replacement: `%%D`},
			{text: `%T`, replacement: `%%T`},
			{text: `%t`, replacement: `%%t`},
			{text: `%@`, replacement: `%%@`},
			{text: `%*`, replacement: `%%*`},
			{text: `%w`, replacement: `%%w`},
			{text: `%W`, replacement: `%%W`},
			{text: `%D`, replacement: `%%D`},
			{text: `%B`, replacement: `%%B`},
			{text: `%b`, replacement: `%%b`},
			{text: `%E`, replacement: `%%E`},
			{text: `%U`, replacement: `%%U`},
			{text: `%S`, replacement: `%%S`},
			{text: `%F`, replacement: `%%F`},
			{text: `%K`, replacement: `%%K`},
			{text: `%G`, replacement: `%%G`},
			{text: `%v`, replacement: `%%v`},
			{text: `%(`, replacement: `%%(`},
		}
	case shell.BASH:
		a.reservedSequences = []sequenceReplacement{
			{text: "`", replacement: "'"},
			{text: `\a`, replacement: `\\a`},
			{text: `\d`, replacement: `\\d`},
			{text: `\D`, replacement: `\\D`},
			{text: `\e`, replacement: `\\e`},
			{text: `\h`, replacement: `\\h`},
			{text: `\H`, replacement: `\\H`},
			{text: `\j`, replacement: `\\j`},
			{text: `\l`, replacement: `\\l`},
			{text: `\n`, replacement: `\\n`},
			{text: `\r`, replacement: `\\r`},
			{text: `\s`, replacement: `\\s`},
			{text: `\t`, replacement: `\\t`},
			{text: `\T`, replacement: `\\T`},
			{text: `\@`, replacement: `\\@`},
			{text: `\A`, replacement: `\\A`},
			{text: `\u`, replacement: `\\u`},
			{text: `\v`, replacement: `\\v`},
			{text: `\V`, replacement: `\\V`},
			{text: `\w`, replacement: `\\w`},
			{text: `\W`, replacement: `\\W`},
			{text: `\!`, replacement: `\\!`},
			{text: `\#`, replacement: `\\#`},
			{text: `\$`, replacement: `\\$`},
			{text: `\nnn`, replacement: `\\nnn`},
		}
	case shell.FISH:
		a.reservedSequences = []sequenceReplacement{
			{text: "`", replacement: "'"},
			{text: `\a`, replacement: `\\a`},
			{text: `\b`, replacement: `\\b`},
			{text: `\e`, replacement: `\\e`},
			{text: `\f`, replacement: `\\f`},
			{text: `\r`, replacement: `\\r`},
			{text: `\t`, replacement: `\\t`},
			{text: `\v`, replacement: `\\v`},
		}
	default:
		a.reservedSequences = []sequenceReplacement{
			{text: "`", replacement: "'"},
		}
	}
}

func (a *Ansi) GenerateHyperlink(text string) string {
	// hyperlink matching
	results := regex.FindNamedRegexMatch("(?P<all>(?:\\[(?P<name>.+)\\])(?:\\((?P<url>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}
	// build hyperlink ansi
	hyperlink := fmt.Sprintf(a.hyperlink, results["url"], results["name"])
	// replace original text by the new one
	return strings.Replace(text, results["all"], hyperlink, 1)
}

func (a *Ansi) formatText(text string) string {
	replaceFormats := func(results []map[string]string) {
		for _, result := range results {
			var formatted string
			switch result["format"] {
			case "b":
				formatted = fmt.Sprintf(a.bold, result["text"])
			case "u":
				formatted = fmt.Sprintf(a.underline, result["text"])
			case "i":
				formatted = fmt.Sprintf(a.italic, result["text"])
			case "s":
				formatted = fmt.Sprintf(a.strikethrough, result["text"])
			case "d":
				formatted = fmt.Sprintf(a.dimmed, result["text"])
			case "f":
				formatted = fmt.Sprintf(a.blink, result["text"])
			case "r":
				formatted = fmt.Sprintf(a.reverse, result["text"])
			}
			text = strings.Replace(text, result["context"], formatted, 1)
		}
	}
	rgx := "(?P<context><(?P<format>[buisrdf])>(?P<text>[^<]+)</[buisrdf]>)"
	for results := regex.FindAllNamedRegexMatch(rgx, text); len(results) != 0; results = regex.FindAllNamedRegexMatch(rgx, text) {
		replaceFormats(results)
	}
	return text
}

func (a *Ansi) CarriageForward() string {
	return fmt.Sprintf(a.right, 1000)
}

func (a *Ansi) GetCursorForRightWrite(length, offset int) string {
	strippedLen := length + (-offset)
	return fmt.Sprintf(a.left, strippedLen)
}

func (a *Ansi) ChangeLine(numberOfLines int) string {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(a.linechange, numberOfLines, position)
}

func (a *Ansi) ConsolePwd(pwd string) string {
	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}
	return fmt.Sprintf(a.osc99, pwd)
}

func (a *Ansi) ClearAfter() string {
	return a.clearLine + a.clearBelow
}

func (a *Ansi) EscapeText(text string) string {
	// what to escape/replace is different per shell

	for _, s := range a.reservedSequences {
		text = strings.ReplaceAll(text, s.text, s.replacement)
	}
	return text
}

func (a *Ansi) Title(title string) string {
	return fmt.Sprintf(a.title, title)
}

func (a *Ansi) ColorReset() string {
	return a.creset
}

func (a *Ansi) FormatText(text string) string {
	return fmt.Sprintf(a.format, text)
}

func (a *Ansi) SaveCursorPosition() string {
	return a.saveCursorPosition
}

func (a *Ansi) RestoreCursorPosition() string {
	return a.restoreCursorPosition
}
