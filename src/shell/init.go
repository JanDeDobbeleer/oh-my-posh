package shell

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

const (
	noExe = "echo \"Unable to find Oh My Posh executable\""
)

var (
	// identify ble.sh by validating the existence of BLE_SESSION_ID
	bashBLEsession bool
)

func getExecutablePath(env runtime.Environment) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	if env.Flags().Strict {
		return path.Base(executable), nil
	}

	// On Windows, it fails when the excutable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	if env.GOOS() == runtime.WINDOWS {
		executable = strings.ReplaceAll(executable, "\\", "/")
	}

	return executable, nil
}

func Init(env runtime.Environment, feats Features) string {
	shell := env.Flags().Shell

	switch shell {
	case ELVISH:
		executable, err := getExecutablePath(env)
		if err != nil {
			return noExe
		}

		var additionalParams string
		if env.Flags().Strict {
			additionalParams += " --strict"
		}

		config := quotePwshOrElvishStr(env.Flags().Config)
		executable = quotePwshOrElvishStr(executable)
		command := "eval ((external %s) init %s --config=%s --print%s | slurp)"

		return fmt.Sprintf(command, executable, shell, config, additionalParams)
	case ZSH, BASH, FISH, CMD, XONSH, NU, PWSH, PWSH5:
		return PrintInit(env, feats, nil)
	default:
		return fmt.Sprintf(`echo "%s is not supported by Oh My Posh"`, shell)
	}
}

func PrintInit(env runtime.Environment, features Features, startTime *time.Time) string {
	shell := env.Flags().Shell
	async := slices.Contains(features, Async)

	if scriptPath, OK := hasScript(env); OK {
		return sourceInit(env, shell, scriptPath, async)
	}

	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	configFile := env.Flags().Config
	sessionID := uuid.NewString()
	bashBLEsession = len(env.Getenv("BLE_SESSION_ID")) != 0

	var script string

	switch shell {
	case PWSH, PWSH5:
		executable = quotePwshOrElvishStr(executable)
		configFile = quotePwshOrElvishStr(configFile)
		sessionID = quotePwshOrElvishStr(sessionID)
		script = pwshInit
	case ZSH:
		executable = QuotePosixStr(executable)
		configFile = QuotePosixStr(configFile)
		sessionID = QuotePosixStr(sessionID)
		script = zshInit
	case BASH:
		executable = QuotePosixStr(executable)
		configFile = QuotePosixStr(configFile)
		sessionID = QuotePosixStr(sessionID)
		script = bashInit
	case FISH:
		executable = quoteFishStr(executable)
		configFile = quoteFishStr(configFile)
		sessionID = quoteFishStr(sessionID)
		script = fishInit
	case CMD:
		executable = escapeLuaStr(executable)
		configFile = escapeLuaStr(configFile)
		sessionID = escapeLuaStr(sessionID)
		script = cmdInit
	case NU:
		executable = quoteNuStr(executable)
		configFile = quoteNuStr(configFile)
		sessionID = quoteNuStr(sessionID)
		script = nuInit
	case ELVISH:
		executable = quotePwshOrElvishStr(executable)
		configFile = quotePwshOrElvishStr(configFile)
		sessionID = quotePwshOrElvishStr(sessionID)
		script = elvishInit
	case XONSH:
		executable = quotePythonStr(executable)
		configFile = quotePythonStr(configFile)
		sessionID = quotePythonStr(sessionID)
		script = xonshInit
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}

	init := strings.NewReplacer(
		"::OMP::", executable,
		"::CONFIG::", configFile,
		"::SHELL::", shell,
		"::SESSION_ID::", sessionID,
	).Replace(script)

	shellScript := features.Lines(shell).String(init)

	log.Debug(shellScript)

	scriptPath, err := writeScript(env, shellScript)
	if err != nil {
		return fmt.Sprintf("echo \"Failed to write init script: %s\"", err.Error())
	}

	if env.Flags().Debug {
		script := sourceInit(env, shell, scriptPath, async)
		log.Debug("init script:", script)
		return printDebug(env, startTime)
	}

	return sourceInit(env, shell, scriptPath, async)
}

func printDebug(env runtime.Environment, startTime *time.Time) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\n%s %s\n", log.Text("Init duration:").Green().Bold().Plain(), time.Since(*startTime)))

	builder.WriteString(log.Text("\n\nLogs:\n\n").Green().Bold().Plain().String())
	builder.WriteString(env.Logs())

	return builder.String()
}

func sourceInit(env runtime.Environment, shell, scriptPath string, async bool) string {
	// nushell stores to autoload, no need to return anything
	if shell == NU {
		return ""
	}

	if env.IsCygwin() {
		var err error
		scriptPath, err = env.RunCommand("cygpath", "-u", scriptPath)
		if err != nil {
			return fmt.Sprintf("echo \"Failed to convert Cygwin path due to %s\"", err.Error())
		}
	}

	if async {
		return sourceInitAsync(shell, scriptPath)
	}

	switch shell {
	case PWSH, PWSH5:
		return scriptPath
	case ZSH:
		return fmt.Sprintf("source '%s'", scriptPath)
	case BASH:
		return fmt.Sprintf("source '%s'", scriptPath)
	case FISH:
		return fmt.Sprintf("source %s", scriptPath)
	case ELVISH:
		return fmt.Sprintf("eval (slurp < %s)", scriptPath)
	case XONSH:
		return fmt.Sprintf("source %s", scriptPath)
	case CMD:
		scriptPath = strings.ReplaceAll(scriptPath, `\`, `\\`)
		return fmt.Sprintf(`load(io.open('%s', "r"):read("*a"))()`, scriptPath)
	default:
		return fmt.Sprintf("echo \"No source command available for %s\"", shell)
	}
}

func sourceInitAsync(shell, scriptPath string) string {
	switch shell {
	case PWSH, PWSH5:
		return fmt.Sprintf(`function prompt() { %s }`, scriptPath)
	case ZSH:
		return fmt.Sprintf(`precmd() { source '%s' }`, scriptPath)
	case BASH:
		return fmt.Sprintf(`PROMPT_COMMAND='source "%s"'`, scriptPath)
	case FISH:
		return fmt.Sprintf(`function fish_prompt; source %s; end`, scriptPath)
	default:
		return ""
	}
}
