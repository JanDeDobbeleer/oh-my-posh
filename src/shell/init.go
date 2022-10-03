package shell

import (
	_ "embed"
	"path/filepath"
	"strconv"

	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/template"
	"os"
	"strings"
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

const (
	noExe = "echo \"Unable to find Oh My Posh executable\""
)

var (
	Transient bool
	ErrorLine bool
	Tooltips  bool
)

func getExecutablePath(env environment.Environment) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	if env.Flags().Strict {
		return environment.Base(env, executable), nil
	}
	// On Windows, it fails when the excutable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	if env.GOOS() == environment.WINDOWS {
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

func Init(env environment.Environment) string {
	shell := env.Flags().Shell
	switch shell {
	case PWSH, PWSH5:
		executable, err := getExecutablePath(env)
		if err != nil {
			return noExe
		}
		if env.Flags().Strict {
			return fmt.Sprintf("(@(& %s init %s --config=%s --print --strict) -join \"`n\") | Invoke-Expression", quotePwshStr(executable), shell, quotePwshStr(env.Flags().Config))
		}
		return fmt.Sprintf("(@(& %s init %s --config=%s --print) -join \"`n\") | Invoke-Expression", quotePwshStr(executable), shell, quotePwshStr(env.Flags().Config))
	case ZSH, BASH, FISH, CMD:
		return PrintInit(env)
	case NU:
		createNuInit(env)
		return ""
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func PrintInit(env environment.Environment) string {
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
	var script string
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
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
	return strings.NewReplacer(
		"::OMP::", executable,
		"::CONFIG::", configFile,
		"::SHELL::", shell,
		"::TRANSIENT::", toggleSetting(Transient),
		"::ERROR_LINE::", toggleSetting(ErrorLine),
		"::TOOLTIPS::", toggleSetting(Tooltips),
	).Replace(script)
}

func createNuInit(env environment.Environment) {
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

func ConsoleBackgroundColor(env environment.Environment, backgroundColorTemplate string) string {
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
