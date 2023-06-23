package shell

import (
	_ "embed"
	"path/filepath"
	"strconv"

	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
)

//go:embed scripts/omp.ps1
var pwshInit string

//go:embed scripts/omp.fish
var fishInit string

//go:embed scripts/omp.bash
var bashInit string

//go:embed scripts/omp.zsh
var zshInit string

//go:embed scripts/omp.lua
var cmdInit string

//go:embed scripts/omp.nu
var nuInit string

//go:embed scripts/omp.tcsh
var tcshInit string

//go:embed scripts/omp.elv
var elvishInit string

//go:embed scripts/omp.py
var xonshInit string

const (
	noExe = "echo \"Unable to find Oh My Posh executable\""
)

var (
	Transient        bool
	ErrorLine        bool
	Tooltips         bool
	ShellIntegration bool
	RPrompt          bool
	Cursor           bool
)

func getExecutablePath(env platform.Environment) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	if env.Flags().Strict {
		return platform.Base(env, executable), nil
	}
	// On Windows, it fails when the excutable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	if env.GOOS() == platform.WINDOWS {
		executable = strings.ReplaceAll(executable, "\\", "/")
	}
	return executable, nil
}

func quotePwshStr(str string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
}

func quotePosixStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	needQuoting := false
	var b strings.Builder
	for _, r := range str {
		normal := false
		switch r {
		case '!', ';', '"', '(', ')', '[', ']', '{', '}', '$', '|', '&', '>', '<', '`', ' ', '#', '~', '*', '?', '=':
			b.WriteRune(r)
		case '\\', '\'':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\v':
			b.WriteString(`\v`)
		default:
			b.WriteRune(r)
			normal = true
		}
		if !normal {
			needQuoting = true
		}
	}
	// the quoting form $'...' is used for a string contains any special characters
	if needQuoting {
		return fmt.Sprintf("$'%s'", b.String())
	}
	return b.String()
}

func quoteFishStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	needQuoting := false
	var b strings.Builder
	for _, r := range str {
		normal := false
		switch r {
		case ';', '"', '(', ')', '[', ']', '{', '}', '$', '|', '&', '>', '<', ' ', '#', '~', '*', '?', '=':
			b.WriteRune(r)
		case '\\', '\'':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
			normal = true
		}
		if !normal {
			needQuoting = true
		}
	}
	// single quotes are used when the string contains any special characters
	if needQuoting {
		return fmt.Sprintf("'%s'", b.String())
	}
	return b.String()
}

func quoteLuaStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	return fmt.Sprintf("'%s'", strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(str))
}

func quoteNuStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	return fmt.Sprintf(`"%s"`, strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(str))
}

func Init(env platform.Environment) string {
	shell := env.Flags().Shell
	switch shell {
	case PWSH, PWSH5, ELVISH:
		executable, err := getExecutablePath(env)
		if err != nil {
			return noExe
		}
		var additionalParams string
		if env.Flags().Strict {
			additionalParams += " --strict"
		}
		if env.Flags().Manual {
			additionalParams += " --manual"
		}
		var command, config string
		switch shell {
		case PWSH, PWSH5:
			command = "(@(& %s init %s --config=%s --print%s) -join \"`n\") | Invoke-Expression"
			config = quotePwshStr(env.Flags().Config)
			executable = quotePwshStr(executable)
		case ELVISH:
			command = "eval (%s init %s --config=%s --print%s | slurp)"
			config = env.Flags().Config
		}
		return fmt.Sprintf(command, executable, shell, config, additionalParams)
	case ZSH, BASH, FISH, CMD, TCSH, XONSH:
		return PrintInit(env)
	case NU:
		createNuInit(env)
		return ""
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func PrintInit(env platform.Environment) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	toggleSetting := func(setting bool) string {
		if env.Flags().Manual {
			return "false"
		}
		return strconv.FormatBool(setting)
	}

	shell := env.Flags().Shell
	configFile := env.Flags().Config

	var (
		script, notice string
		hasNotice      bool
	)

	switch shell {
	case PWSH, PWSH5:
		executable = quotePwshStr(executable)
		configFile = quotePwshStr(configFile)
		script = pwshInit
	case ZSH:
		executable = quotePosixStr(executable)
		configFile = quotePosixStr(configFile)
		script = zshInit
	case BASH:
		executable = quotePosixStr(executable)
		configFile = quotePosixStr(configFile)
		script = bashInit
	case FISH:
		executable = quoteFishStr(executable)
		configFile = quoteFishStr(configFile)
		script = fishInit
	case CMD:
		executable = quoteLuaStr(executable)
		configFile = quoteLuaStr(configFile)
		script = cmdInit
	case NU:
		executable = quoteNuStr(executable)
		configFile = quoteNuStr(configFile)
		script = nuInit
	case TCSH:
		executable = quotePosixStr(executable)
		configFile = quotePosixStr(configFile)
		script = tcshInit
	case ELVISH:
		script = elvishInit
	case XONSH:
		script = xonshInit
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}

	// only run this for shells that support
	// injecting the notice directly
	if shell != PWSH && shell != PWSH5 {
		notice, hasNotice = upgrade.Notice(env)
	}

	return strings.NewReplacer(
		"::OMP::", executable,
		"::CONFIG::", configFile,
		"::SHELL::", shell,
		"::TRANSIENT::", toggleSetting(Transient),
		"::ERROR_LINE::", toggleSetting(ErrorLine),
		"::TOOLTIPS::", toggleSetting(Tooltips),
		"::FTCS_MARKS::", toggleSetting(ShellIntegration),
		"::RPROMPT::", strconv.FormatBool(RPrompt),
		"::CURSOR::", strconv.FormatBool(Cursor),
		"::UPGRADE::", strconv.FormatBool(hasNotice),
		"::UPGRADENOTICE::", notice,
	).Replace(script)
}

func createNuInit(env platform.Environment) {
	initPath := filepath.Join(env.Home(), ".oh-my-posh.nu")
	f, err := os.OpenFile(initPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	_, err = f.WriteString(PrintInit(env))
	if err != nil {
		return
	}
	_ = f.Close()
}

func ConsoleBackgroundColor(env platform.Environment, backgroundColorTemplate string) string {
	if len(backgroundColorTemplate) == 0 {
		return backgroundColorTemplate
	}
	tmpl := &template.Text{
		Template: backgroundColorTemplate,
		Context:  nil,
		Env:      env,
	}
	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}
	return text
}
