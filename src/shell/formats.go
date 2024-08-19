package shell

type Formats struct {
	Escape     string
	Left       string
	Linechange string
	ClearBelow string
	ClearLine  string

	Title string

	SaveCursorPosition    string
	RestoreCursorPosition string

	Osc99 string
	Osc7  string
	Osc51 string

	EscapeSequences map[rune]rune

	HyperlinkStart  string
	HyperlinkCenter string
	HyperlinkEnd    string

	ITermPromptMark string
	ITermCurrentDir string
	ITermRemoteHost string
}

func GetFormats(shell string) *Formats {
	var formats *Formats

	switch shell {
	case BASH:
		formats = &Formats{
			Escape:                "\\[%s\\]",
			Linechange:            "\\[\x1b[%d%s\\]",
			Left:                  "\\[\x1b[%dD\\]",
			ClearBelow:            "\\[\x1b[0J\\]",
			ClearLine:             "\\[\x1b[K\\]",
			SaveCursorPosition:    "\\[\x1b7\\]",
			RestoreCursorPosition: "\\[\x1b8\\]",
			Title:                 "\\[\x1b]0;%s\007\\]",
			HyperlinkStart:        "\\[\x1b]8;;",
			HyperlinkCenter:       "\x1b\\\\\\]",
			HyperlinkEnd:          "\\[\x1b]8;;\x1b\\\\\\]",
			Osc99:                 "\\[\x1b]9;9;%s\x1b\\\\\\]",
			Osc7:                  "\\[\x1b]7;file://%s/%s\x1b\\\\\\]",
			Osc51:                 "\\[\x1b]51;A;%s@%s:%s\x1b\\\\\\]",
			ITermPromptMark:       "\\[$(iterm2_prompt_mark)\\]",
			ITermCurrentDir:       "\\[\x1b]1337;CurrentDir=%s\x07\\]",
			ITermRemoteHost:       "\\[\x1b]1337;RemoteHost=%s@%s\x07\\]",
			EscapeSequences: map[rune]rune{
				96: 92, // backtick
				92: 92, // backslash
			},
		}
	case ZSH, TCSH:
		formats = &Formats{
			Escape:                "%%{%s%%}",
			Linechange:            "%%{\x1b[%d%s%%}",
			Left:                  "%%{\x1b[%dD%%}",
			ClearBelow:            "%{\x1b[0J%}",
			ClearLine:             "%{\x1b[K%}",
			SaveCursorPosition:    "%{\x1b7%}",
			RestoreCursorPosition: "%{\x1b8%}",
			Title:                 "%%{\x1b]0;%s\007%%}",
			HyperlinkStart:        "%{\x1b]8;;",
			HyperlinkCenter:       "\x1b\\%}",
			HyperlinkEnd:          "%{\x1b]8;;\x1b\\%}",
			Osc99:                 "%%{\x1b]9;9;%s\x1b\\%%}",
			Osc7:                  "%%{\x1b]7;file://%s/%s\x1b\\%%}",
			Osc51:                 "%%{\x1b]51;A%s@%s:%s\x1b\\%%}",
			ITermPromptMark:       "%{$(iterm2_prompt_mark)%}",
			ITermCurrentDir:       "%%{\x1b]1337;CurrentDir=%s\x07%%}",
			ITermRemoteHost:       "%%{\x1b]1337;RemoteHost=%s@%s\x07%%}",
		}
	default:
		formats = &Formats{
			Escape:                "%s",
			Linechange:            "\x1b[%d%s",
			Left:                  "\x1b[%dD",
			ClearBelow:            "\x1b[0J",
			ClearLine:             "\x1b[K",
			SaveCursorPosition:    "\x1b7",
			RestoreCursorPosition: "\x1b8",
			Title:                 "\x1b]0;%s\007",
			// when in fish on Linux, it seems hyperlinks ending with \\ print a \
			// unlike on macOS. However, this is a fish bug, so do not try to fix it here:
			// https://github.com/JanDeDobbeleer/oh-my-posh/pull/3288#issuecomment-1369137068
			HyperlinkStart:  "\x1b]8;;",
			HyperlinkCenter: "\x1b\\",
			HyperlinkEnd:    "\x1b]8;;\x1b\\",
			Osc99:           "\x1b]9;9;%s\x1b\\",
			Osc7:            "\x1b]7;file://%s/%s\x1b\\",
			Osc51:           "\x1b]51;A%s@%s:%s\x1b\\",
			ITermCurrentDir: "\x1b]1337;CurrentDir=%s\x07",
			ITermRemoteHost: "\x1b]1337;RemoteHost=%s@%s\x07",
		}
	}

	if shell == ZSH {
		formats.EscapeSequences = map[rune]rune{
			96: 92, // backtick
			37: 37, // %
		}
	}

	return formats
}
