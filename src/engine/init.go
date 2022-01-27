package engine

import (
	_ "embed"

	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/template"
	"os"
	"strings"
)

//go:embed init/omp.ps1
var pwshInit string

//go:embed init/omp.fish
var fishInit string

//go:embed init/omp.bash
var bashInit string

//go:embed init/omp.zsh
var zshInit string

//go:embed init/omp.lua
var cmdInit string

const (
	noExe = "echo \"Unable to find Oh My Posh executable\""

	zsh         = "zsh"
	bash        = "bash"
	pwsh        = "pwsh"
	fish        = "fish"
	powershell5 = "powershell"
	winCMD      = "cmd"
	plain       = "shell"
)

func getExecutablePath(shell string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	// On Windows, it fails when the excutable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	executable = strings.ReplaceAll(executable, "\\", "/")
	switch shell {
	case bash, zsh:
		return strings.ReplaceAll(executable, " ", "\\ "), nil
	}
	return executable, nil
}

func InitShell(shell, configFile string) string {
	executable, err := getExecutablePath(shell)
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh, powershell5:
		return fmt.Sprintf("(@(&\"%s\" --print-init --shell=%s --config=\"%s\") -join \"`n\") | Invoke-Expression", executable, shell, configFile)
	case zsh, bash, fish, winCMD:
		return PrintShellInit(shell, configFile)
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func PrintShellInit(shell, configFile string) string {
	executable, err := getExecutablePath(shell)
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh, powershell5:
		return getShellInitScript(executable, configFile, pwshInit)
	case zsh:
		return getShellInitScript(executable, configFile, zshInit)
	case bash:
		return getShellInitScript(executable, configFile, bashInit)
	case fish:
		return getShellInitScript(executable, configFile, fishInit)
	case winCMD:
		return getShellInitScript(executable, configFile, cmdInit)
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func getShellInitScript(executable, configFile, script string) string {
	script = strings.ReplaceAll(script, "::OMP::", executable)
	script = strings.ReplaceAll(script, "::CONFIG::", configFile)
	return script
}

func GetConsoleBackgroundColor(env environment.Environment, backgroundColorTemplate string) string {
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
