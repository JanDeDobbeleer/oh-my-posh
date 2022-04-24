package shell

import (
	_ "embed"
	"path/filepath"

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
	executable = strings.ReplaceAll(executable, "\\", "/")
	switch env.Flags().Shell {
	case BASH, ZSH:
		executable = strings.ReplaceAll(executable, " ", "\\ ")
		executable = strings.ReplaceAll(executable, "(", "\\(")
		executable = strings.ReplaceAll(executable, ")", "\\)")
	}
	return executable, nil
}

func Init(env environment.Environment) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}
	shell := env.Flags().Shell
	switch shell {
	case PWSH, PWSH5:
		return fmt.Sprintf("(@(&\"%s\" init %s --config=\"%s\" --print) -join \"`n\") | Invoke-Expression", executable, shell, env.Flags().Config)
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
	shell := env.Flags().Shell
	configFile := env.Flags().Config
	switch shell {
	case PWSH, PWSH5:
		return getShellInitScript(executable, configFile, pwshInit)
	case ZSH:
		return getShellInitScript(executable, configFile, zshInit)
	case BASH:
		return getShellInitScript(executable, configFile, bashInit)
	case FISH:
		return getShellInitScript(executable, configFile, fishInit)
	case CMD:
		return getShellInitScript(executable, configFile, cmdInit)
	case NU:
		return getShellInitScript(executable, configFile, nuInit)
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func getShellInitScript(executable, configFile, script string) string {
	script = strings.ReplaceAll(script, "::OMP::", executable)
	script = strings.ReplaceAll(script, "::CONFIG::", configFile)
	return script
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
