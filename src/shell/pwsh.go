package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.ps1
var pwshInit string

func (f Feature) Pwsh() Code {
	switch f {
	case Tooltips:
		return "Enable-PoshTooltips"
	case LineError:
		return "Enable-PoshLineError"
	case Transient:
		return "Enable-PoshTransientPrompt"
	case Jobs:
		return "$global:_ompJobCount = $true"
	case Azure:
		return "$global:_ompAzure = $true"
	case PoshGit:
		return "$global:_ompPoshGit = $true"
	case FTCSMarks:
		return "$global:_ompFTCSMarks = $true"
	case Upgrade:
		return "& $global:_ompExecutable upgrade"
	case Notice:
		return "& $global:_ompExecutable notice"
	case PromptMark, RPrompt, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func quotePwshOrElvishStr(str string) string {
	if len(str) == 0 {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
}
