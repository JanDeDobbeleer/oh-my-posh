package shell

import (
	_ "embed"

	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

//go:embed scripts/omp.nu
var nuInit string

func (f Feature) Nu() Code {
	switch f {
	case Transient:
		return `$env.TRANSIENT_PROMPT_COMMAND = { ^$_omp_executable print transient $"--config=($env.POSH_THEME)" --shell=nu $"--shell-version=($env.POSH_SHELL_VERSION)" $"--execution-time=(posh_cmd_duration)" $"--status=($env.LAST_EXIT_CODE)" $"--terminal-width=(posh_width)" }` //nolint: lll
	case Upgrade:
		return "^$_omp_executable upgrade"
	case Notice:
		return "^$_omp_executable notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, FTCSMarks, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func quoteNuStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	return fmt.Sprintf(`"%s"`, strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(str))
}

func createNuInit(env runtime.Environment, features Features) {
	initPath := filepath.Join(env.Home(), ".oh-my-posh.nu")
	f, err := os.OpenFile(initPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}

	_, err = f.WriteString(PrintInit(env, features))
	if err != nil {
		return
	}

	_ = f.Close()
}
