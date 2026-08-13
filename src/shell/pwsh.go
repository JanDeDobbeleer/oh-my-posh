package shell

import (
	_ "embed"
	"fmt"
	"strings"
	"unicode/utf8"
)

//go:embed scripts/omp.ps1
var pwshInit string

func (f Features) Pwsh() Code {
	switch f {
	case Tooltips:
		return "Enable-PoshTooltips"
	case LineError:
		return "Enable-PoshLineError"
	case Transient:
		return "$global:_ompTransientPrompt = $true"
	case Jobs:
		return "$global:_ompJobCount = $true"
	case Azure:
		return "$global:_ompAzure = $true"
	case PoshGit:
		return "$global:_ompPoshGit = $true"
	case FTCSMarks:
		return "$global:_ompFTCSMarks = $true"
	case Upgrade:
		return "& $global:_ompExecutable upgrade --auto"
	case Notice:
		return "& $global:_ompExecutable notice"
	case Streaming:
		return "Enable-PoshStreaming"
	case KeyHandlers:
		return "Enable-KeyHandlers"
	case VIMode:
		return "Enable-PoshVIMode"
	case PromptMark, RPrompt, CursorPositioning, Async:
		fallthrough
	default:
		return ""
	}
}

func quotePwshStr(str string) string {
	if str == "" {
		return "''"
	}

	ascii := strings.IndexFunc(str, func(r rune) bool { return r >= utf8.RuneSelf }) == -1
	if ascii {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
	}

	// PowerShell decodes native command output using [Console]::OutputEncoding,
	// which defaults to the legacy OEM code page on Windows. Multi-byte UTF-8
	// sequences in stdout are mangled before the init script gets a chance to
	// run, so non-ASCII runes must be spelled out as [char] expressions to keep
	// the emitted code pure ASCII, which survives every code page.
	// The expandable string form "$( )" stays a single token in argument mode,
	// so the result can be used anywhere a quoted string can.
	var parts []string

	var segment strings.Builder

	flush := func() {
		if segment.Len() == 0 {
			return
		}

		parts = append(parts, fmt.Sprintf("'%s'", strings.ReplaceAll(segment.String(), "'", "''")))
		segment.Reset()
	}

	for _, r := range str {
		if r < utf8.RuneSelf {
			segment.WriteRune(r)
			continue
		}

		flush()

		if r > 0xFFFF {
			parts = append(parts, fmt.Sprintf("[char]::ConvertFromUtf32(0x%X)", r))
			continue
		}

		parts = append(parts, fmt.Sprintf("[char]0x%X", r))
	}

	flush()

	// a leading [char] would make + do character arithmetic instead of
	// string concatenation, an empty string first forces string semantics
	if !strings.HasPrefix(parts[0], "'") {
		parts = append([]string{"''"}, parts...)
	}

	return fmt.Sprintf(`"$(%s)"`, strings.Join(parts, " + "))
}
