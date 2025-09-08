package shell

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
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
	switch env.Flags().Shell {
	case ELVISH, PWSH, PWSH5:
		if env.Flags().Shell != ELVISH && !env.Flags().Eval {
			return PrintInit(env, feats, nil)
		}

		executable, err := getExecutablePath(env)
		if err != nil {
			return noExe
		}

		var additionalParams string
		if env.Flags().Strict {
			additionalParams += " --strict"
		}

		config := quotePwshOrElvishStr(env.Flags().ConfigPath)
		executable = quotePwshOrElvishStr(executable)

		var command string

		switch env.Flags().Shell {
		case PWSH, PWSH5:
			command = "(@(& %s init %s --config=%s --print --eval%s) -join \"`n\") | Invoke-Expression"
		case ELVISH:
			command = "eval ((external %s) init %s --config=%s --print%s | slurp)"
		}

		return fmt.Sprintf(command, executable, env.Flags().Shell, config, additionalParams)
	case ZSH, BASH, FISH, CMD, XONSH, NU:
		return PrintInit(env, feats, nil)
	default:
		return fmt.Sprintf(`echo "%s is not supported by Oh My Posh"`, env.Flags().Shell)
	}
}

func PrintInit(env runtime.Environment, features Features, startTime *time.Time) string {
	async := features&Async != 0

	if scriptPath, OK := hasScript(env); OK {
		return sourceInit(env, scriptPath, async)
	}

	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	bashBLEsession = len(env.Getenv("BLE_SESSION_ID")) != 0

	var script string

	switch env.Flags().Shell {
	case PWSH, PWSH5:
		executable = quotePwshOrElvishStr(executable)
		script = pwshInit
	case ZSH:
		executable = QuotePosixStr(executable)
		script = zshInit
	case BASH:
		executable = QuotePosixStr(executable)
		script = bashInit
	case FISH:
		executable = quoteFishStr(executable)
		script = fishInit
	case CMD:
		executable = escapeLuaStr(executable)
		script = cmdInit
	case NU:
		executable = quoteNuStr(executable)
		script = nuInit
	case ELVISH:
		executable = quotePwshOrElvishStr(executable)
		script = elvishInit
	case XONSH:
		executable = quotePythonStr(executable)
		script = xonshInit
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", env.Flags().Shell)
	}

	init := strings.NewReplacer(
		"::OMP::", executable,
		"::SHELL::", env.Flags().Shell,
		"::SESSION_ID::", cache.SessionID(),
	).Replace(script)

	shellScript := features.Lines(env.Flags().Shell).String(init)

	if env.Flags().Eval {
		return shellScript
	}

	log.Debug(shellScript)

	scriptPath, err := writeScript(env, shellScript)
	if err != nil {
		return fmt.Sprintf("echo \"Failed to write init script: %s\"", err.Error())
	}

	sourceCommand := sourceInit(env, scriptPath, async)

	if !env.Flags().Debug {
		return sourceCommand
	}

	if len(sourceCommand) != 0 {
		log.Debug("init source command:", sourceCommand)
	}

	return printDebug(env, startTime)
}

func printDebug(env runtime.Environment, startTime *time.Time) string {
	builder := text.NewBuilder()

	builder.WriteString(fmt.Sprintf("\n%s %s\n", log.Text("Init duration:").Green().Bold().Plain(), time.Since(*startTime)))

	builder.WriteString(log.Text("\n\nLogs:\n\n").Green().Bold().Plain().String())
	builder.WriteString(env.Logs())

	return builder.String()
}

func sourceInit(env runtime.Environment, scriptPath string, async bool) string {
	// nushell stores to autoload, no need to return anything
	if env.Flags().Shell == NU {
		return ""
	}

	if env.IsCygwin() {
		var err error
		scriptPath, err = env.RunCommand("cygpath", "-u", scriptPath)
		if err != nil {
			log.Error(err)
			return fmt.Sprintf("echo \"Failed to convert Cygwin path due to %s\"", err.Error())
		}
	}

	script := sessionScript(env.Flags().Shell)

	if async {
		return script + sourceInitAsync(env.Flags().Shell, scriptPath)
	}

	switch env.Flags().Shell {
	case PWSH, PWSH5:
		script += fmt.Sprintf("& %s", quotePwshOrElvishStr(scriptPath))
	case ZSH, BASH:
		script += fmt.Sprintf("source %s", QuotePosixStr(scriptPath))
	case XONSH:
		script += fmt.Sprintf("source %s", quotePythonStr(scriptPath))
	case FISH:
		script += fmt.Sprintf("source %s", quoteFishStr(scriptPath))
	case ELVISH:
		script += fmt.Sprintf("eval (slurp < %s)", quotePwshOrElvishStr(scriptPath))
	case CMD:
		script += fmt.Sprintf(`load(io.open('%s', "r"):read("*a"))()`, escapeLuaStr(scriptPath))
	default:
		return fmt.Sprintf("echo \"No source command available for %s\"", env.Flags().Shell)
	}

	return script
}

func sourceInitAsync(shell, scriptPath string) string {
	switch shell {
	case PWSH, PWSH5:
		return fmt.Sprintf("function prompt() { & %s }", quotePwshOrElvishStr(scriptPath))
	case ZSH:
		return fmt.Sprintf("precmd() { source %s }", QuotePosixStr(scriptPath))
	case BASH:
		command := fmt.Sprintf("source %s", QuotePosixStr(scriptPath))
		return fmt.Sprintf("PROMPT_COMMAND=%s", QuotePosixStr(command))
	case FISH:
		return fmt.Sprintf("function fish_prompt; source %s; end", quoteFishStr(scriptPath))
	default:
		return ""
	}
}

func sessionScript(shell string) string {
	switch shell {
	case PWSH, PWSH5:
		return fmt.Sprintf("$env:POSH_SESSION_ID = \"%s\";", cache.SessionID())
	case ZSH, BASH:
		return fmt.Sprintf("export POSH_SESSION_ID=\"%s\";", cache.SessionID())
	case XONSH:
		return fmt.Sprintf("$POSH_SESSION_ID = \"%s\";", cache.SessionID())
	case FISH:
		return fmt.Sprintf("set --export --global POSH_SESSION_ID \"%s\";", cache.SessionID())
	case ELVISH:
		return fmt.Sprintf("set-env POSH_SESSION_ID \"%s\";", cache.SessionID())
	case CMD:
		return fmt.Sprintf(`os.setenv('POSH_SESSION_ID', '%s');`, cache.SessionID())
	}
	return ""
}
