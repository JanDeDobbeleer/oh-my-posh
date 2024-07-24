package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
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
		return runtime.Base(env, executable), nil
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
		return PrintInit(env, feats)
	case NU:
		createNuInit(env, feats)
		return ""
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func PrintInit(env runtime.Environment, features Features) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
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

	init := strings.NewReplacer(
		"::OMP::", executable,
		"::CONFIG::", configFile,
		"::SHELL::", shell,
	).Replace(script)

	return features.Lines(shell).String(init)
}
