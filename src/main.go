package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/config/v2"
)

// Version number of oh-my-posh
var Version = "development"

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
	noExe       = "echo \"Unable to find Oh My Posh executable\""
	zsh         = "zsh"
	bash        = "bash"
	pwsh        = "pwsh"
	fish        = "fish"
	powershell5 = "powershell"
	winCMD      = "cmd"
	plain       = "shell"
)

type args struct {
	ErrorCode      *int
	PrintConfig    *bool
	ConfigFormat   *string
	PrintShell     *bool
	Config         *string
	Shell          *string
	PWD            *string
	PSWD           *string
	Version        *bool
	Debug          *bool
	ExecutionTime  *float64
	Millis         *bool
	Eval           *bool
	Init           *bool
	PrintInit      *bool
	ExportPNG      *bool
	Author         *string
	CursorPadding  *int
	RPromptOffset  *int
	RPrompt        *bool
	BGColor        *string
	StackCount     *int
	Command        *string
	PrintTransient *bool
	Plain          *bool
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
		ConfigFormat: flag.String(
			"config-format",
			config.JSON,
			"The format to print the config in. Valid options are:\n- json\n- yaml\n- toml\n"),
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
		PSWD: flag.String(
			"pswd",
			"",
			"the powershell path you are working in, useful when working with drives"),
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
		ExportPNG: flag.Bool(
			"export-png",
			false,
			"Create an image based on the current configuration"),
		Author: flag.String(
			"author",
			"",
			"Add the author to the exported image using --export-img"),
		CursorPadding: flag.Int(
			"cursor-padding",
			30,
			"Pad the cursor with x when using --export-img"),
		RPromptOffset: flag.Int(
			"rprompt-offset",
			40,
			"Offset the right prompt with x when using --export-img"),
		RPrompt: flag.Bool(
			"rprompt",
			false,
			"Only print the rprompt block"),
		BGColor: flag.String(
			"bg-color",
			"#151515",
			"Set the background color when using --export-img"),
		StackCount: flag.Int(
			"stack-count",
			0,
			"The current location stack count"),
		Command: flag.String(
			"command",
			"",
			"Render a tooltip based on the command value"),
		PrintTransient: flag.Bool(
			"print-transient",
			false,
			"Print the transient prompt"),
		Plain: flag.Bool(
			"plain",
			false,
			"Print a plain prompt without ANSI"),
	}
	flag.Parse()
	env := &environment{}
	env.init(args)
	defer env.close()
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
	if *args.PrintConfig {
		fmt.Print(exportConfig(*args.Config, *args.ConfigFormat))
		return
	}
	cfg := GetConfig(env)
	if *args.PrintShell {
		fmt.Println(env.getShellName())
		return
	}
	if *args.Version {
		fmt.Println(Version)
		return
	}

	ansi := &ansiUtils{}
	ansi.init(env.getShellName())
	var writer promptWriter
	if *args.Plain {
		writer = &PlainWriter{}
	} else {
		writerColors := MakeColors(env, cfg)
		writer = &AnsiWriter{
			ansi:               ansi,
			terminalBackground: getConsoleBackgroundColor(env, cfg.TerminalBackground),
			ansiColors:         writerColors,
		}
	}
	title := &consoleTitle{
		env:    env,
		config: cfg,
		ansi:   ansi,
	}
	engine := &engine{
		config:       cfg,
		env:          env,
		writer:       writer,
		consoleTitle: title,
		ansi:         ansi,
		plain:        *args.Plain,
	}
	if *args.Debug {
		fmt.Print(engine.debug())
		return
	}
	if *args.PrintTransient {
		fmt.Print(engine.renderTransientPrompt())
		return
	}
	if len(*args.Command) != 0 {
		fmt.Print(engine.renderTooltip(*args.Command))
		return
	}
	if *args.RPrompt {
		fmt.Print(engine.renderRPrompt())
		return
	}
	prompt := engine.render()
	if !*args.ExportPNG {
		fmt.Print(prompt)
		return
	}
	imageCreator := &ImageRenderer{
		ansiString:    prompt,
		author:        *args.Author,
		cursorPadding: *args.CursorPadding,
		rPromptOffset: *args.RPromptOffset,
		bgColor:       *args.BGColor,
		ansi:          ansi,
	}
	imageCreator.init()
	match := findNamedRegexMatch(`.*(\/|\\)(?P<STR>.+).omp.(json|yaml|toml)`, *args.Config)
	err := imageCreator.SavePNG(fmt.Sprintf("%s.png", match[str]))
	if err != nil {
		fmt.Print(err.Error())
	}
}

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

func initShell(shell, configFile string) string {
	executable, err := getExecutablePath(shell)
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh:
		return fmt.Sprintf("(@(&\"%s\" --print-init --shell=pwsh --config=\"%s\") -join \"`n\") | Invoke-Expression", executable, configFile)
	case zsh, bash, fish, winCMD:
		return printShellInit(shell, configFile)
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", shell)
	}
}

func printShellInit(shell, configFile string) string {
	executable, err := getExecutablePath(shell)
	if err != nil {
		return noExe
	}
	switch shell {
	case pwsh:
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

func getConsoleBackgroundColor(env environmentInfo, backgroundColorTemplate string) string {
	if len(backgroundColorTemplate) == 0 {
		return backgroundColorTemplate
	}
	context := struct {
		Env map[string]string
	}{
		Env: map[string]string{},
	}
	matches := findAllNamedRegexMatch(templateEnvRegex, backgroundColorTemplate)
	for _, match := range matches {
		context.Env[match["ENV"]] = env.getenv(match["ENV"])
	}
	template := &textTemplate{
		Template: backgroundColorTemplate,
		Context:  context,
		Env:      env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}
