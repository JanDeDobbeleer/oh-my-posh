//go:generate go-bindata -o init.go init/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Version number of oh-my-posh
var Version = "development"

const (
	noExe       = "echo \"Unable to find Oh my Posh executable\""
	zsh         = "zsh"
	bash        = "bash"
	pwsh        = "pwsh"
	fish        = "fish"
	powershell5 = "powershell"
)

type args struct {
	ErrorCode     *int
	PrintConfig   *bool
	PrintShell    *bool
	Config        *string
	Shell         *string
	PWD           *string
	Version       *bool
	Debug         *bool
	ExecutionTime *float64
	Millis        *bool
	Eval          *bool
	Init          *bool
	PrintInit     *bool
}

func main() {
	args := &args{
		ErrorCode: flag.Int(
			"error",
			0,
			"Error code of previously executed command"),
		PrintConfig: flag.Bool(
			"print-config",
			false,
			"Print the current config in json format"),
		PrintShell: flag.Bool(
			"print-shell",
			false,
			"Print the current shell name"),
		Config: flag.String(
			"config",
			"",
			"Add the path to a configuration you wish to load"),
		Shell: flag.String(
			"shell",
			"",
			"Override the shell you are working in"),
		PWD: flag.String(
			"pwd",
			"",
			"the path you are working in"),
		Version: flag.Bool(
			"version",
			false,
			"Print the current version of the binary"),
		Debug: flag.Bool(
			"debug",
			false,
			"Print debug information"),
		ExecutionTime: flag.Float64(
			"execution-time",
			0,
			"Execution time of the previously executed command"),
		Millis: flag.Bool(
			"millis",
			false,
			"Get the current time in milliseconds"),
		Eval: flag.Bool(
			"eval",
			false,
			"Run in eval mode"),
		Init: flag.Bool(
			"init",
			false,
			"Initialize the shell"),
		PrintInit: flag.Bool(
			"print-init",
			false,
			"Print the shell initialization script"),
	}
	flag.Parse()
	env := &environment{
		args:     args,
		commands: make(map[string]string),
		cwd:      *args.PWD,
	}
	if *args.Millis {
		fmt.Print(time.Now().UnixNano() / 1000000)
		return
	}
	if *args.Init {
		init := initShell(*args.Shell, *args.Config)
		fmt.Print(init)
		return
	}
	if *args.PrintInit {
		init := printShellInit(*args.Shell, *args.Config)
		fmt.Print(init)
		return
	}
	settings := GetSettings(env)
	if *args.PrintConfig {
		theme, _ := json.MarshalIndent(settings, "", "    ")
		fmt.Println(string(theme))
		return
	}
	if *args.PrintShell {
		fmt.Println(env.getShellName())
		return
	}
	if *args.Version {
		fmt.Println(Version)
		return
	}

	formats := &ansiFormats{}
	formats.init(env.getShellName())

	renderer := &AnsiRenderer{
		buffer:  new(bytes.Buffer),
		formats: formats,
	}
	colorer := &AnsiColor{
		buffer:  new(bytes.Buffer),
		formats: formats,
	}
	title := &consoleTitle{
		env:      env,
		settings: settings,
		formats:  formats,
	}
	engine := &engine{
		settings:     settings,
		env:          env,
		color:        colorer,
		renderer:     renderer,
		consoleTitle: title,
	}

	if *args.Debug {
		engine.debug()
		return
	}
	engine.render()
}

func initShell(shell, config string) string {
	executable, err := os.Executable()
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh:
		return fmt.Sprintf("Invoke-Expression (@(&\"%s\" --print-init --shell pwsh --config %s) -join \"`n\")", executable, config)
	case zsh, bash, fish:
		return printShellInit(shell, config)
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func printShellInit(shell, config string) string {
	executable, err := os.Executable()
	// On Windows, it fails when the excutable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	executable = strings.ReplaceAll(executable, "\\", "/")
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh:
		return getShellInitScript(executable, config, "init/omp.ps1")
	case zsh:
		return getShellInitScript(executable, config, "init/omp.zsh")
	case bash:
		return getShellInitScript(executable, config, "init/omp.bash")
	case fish:
		return getShellInitScript(executable, config, "init/omp.fish")
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func getShellInitScript(executable, config, script string) string {
	data, err := Asset(script)
	if err != nil {
		return fmt.Sprintf("echo \"Unable to find initialization script %s\"", script)
	}
	init := string(data)
	init = strings.ReplaceAll(init, "::OMP::", executable)
	init = strings.ReplaceAll(init, "::CONFIG::", config)
	return init
}
