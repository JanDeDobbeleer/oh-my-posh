package shell

import (
	"fmt"
	"os"
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
	case PWSH, PWSH5, ELVISH:
		executable, err := getExecutablePath(env)
		if err != nil {
			return noExe
		}

		var additionalParams string
		if env.Flags().Strict {
			additionalParams += " --strict"
		}

		var command, config string

		switch shell {
		case PWSH, PWSH5:
			command = "(@(& %s init %s --config=%s --print%s) -join \"`n\") | Invoke-Expression"
		case ELVISH:
			command = "eval ((external %s) init %s --config=%s --print%s | slurp)"
		}

		config = quotePwshOrElvishStr(env.Flags().Config)
		executable = quotePwshOrElvishStr(executable)

		return fmt.Sprintf(command, executable, shell, config, additionalParams)
	case ZSH, BASH, FISH, CMD, TCSH, XONSH:
		return PrintInit(env, feats, nil)
	case NU:
		createNuInit(env, feats)
		return ""
	default:
		return fmt.Sprintf(`echo "%s is not supported by Oh My Posh"`, shell)
	}
}

func PrintInit(env runtime.Environment, features Features, startTime *time.Time) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	shell := env.Flags().Shell
	configFile := env.Flags().Config
	sessionID := uuid.NewString()

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
	case TCSH:
		executable = quoteCshStr(executable)
		configFile = quoteCshStr(configFile)
		sessionID = quoteCshStr(sessionID)
		script = tcshInit
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

	if !env.Flags().Debug {
		return shellScript
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\n%s %s\n", log.Text("Init duration:").Green().Bold().Plain(), time.Since(*startTime)))

	builder.WriteString(log.Text("\n\nLogs:\n\n").Green().Bold().Plain().String())
	builder.WriteString(env.Logs())

	return builder.String()
}
