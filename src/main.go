package main

import (
	"flag"
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/engine"
	"oh-my-posh/environment"
	"time"

	"github.com/gookit/config/v2"
)

// Version number of oh-my-posh
var Version = "development"

func main() {
	args := &environment.Args{
		ErrorCode: flag.Int(
			"error",
			0,
			"Error code of previously executed command"),
		PrintConfig: flag.Bool(
			"print-config",
			false,
			"Print the current config in json format"),
		ConfigFormat: flag.String(
			"format",
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
		CachePath: flag.Bool(
			"cache-path",
			false,
			"Print the location of the cache"),
		Migrate: flag.Bool(
			"migrate",
			false,
			"Migrate the config to the latest version"),
		Write: flag.Bool(
			"write",
			false,
			"Write the config to the file"),
	}
	flag.Parse()
	if *args.Version {
		fmt.Println(Version)
		return
	}
	env := &environment.ShellEnvironment{}
	env.Init(args)
	defer env.Close()
	if *args.PrintShell {
		fmt.Println(env.Shell())
		return
	}
	if *args.Millis {
		fmt.Print(time.Now().UnixNano() / 1000000)
		return
	}
	if *args.CachePath {
		fmt.Print(env.CachePath())
		return
	}
	if *args.Init {
		init := engine.InitShell(*args.Shell, *args.Config)
		fmt.Print(init)
		return
	}
	if *args.PrintInit {
		init := engine.PrintShellInit(*args.Shell, *args.Config)
		fmt.Print(init)
		return
	}
	cfg := engine.LoadConfig(env)
	if *args.PrintConfig {
		fmt.Print(cfg.Export(*args.ConfigFormat))
		return
	}
	if *args.Migrate {
		if *args.Write {
			cfg.BackupAndMigrate(env)
			return
		}
		cfg.Migrate(env)
		fmt.Print(cfg.Export(*args.ConfigFormat))
		return
	}
	ansi := &color.Ansi{}
	ansi.Init(env.Shell())
	var writer color.Writer
	if *args.Plain {
		writer = &color.PlainWriter{}
	} else {
		writerColors := cfg.MakeColors(env)
		writer = &color.AnsiWriter{
			Ansi:               ansi,
			TerminalBackground: engine.GetConsoleBackgroundColor(env, cfg.TerminalBackground),
			AnsiColors:         writerColors,
		}
	}
	consoleTitle := &console.Title{
		Env:      env,
		Ansi:     ansi,
		Template: cfg.ConsoleTitleTemplate,
		Style:    cfg.ConsoleTitleStyle,
	}
	eng := &engine.Engine{
		Config:       cfg,
		Env:          env,
		Writer:       writer,
		ConsoleTitle: consoleTitle,
		Ansi:         ansi,
		Plain:        *args.Plain,
	}
	if *args.Debug {
		fmt.Print(eng.Debug(Version))
		return
	}
	if *args.PrintTransient {
		fmt.Print(eng.RenderTransientPrompt())
		return
	}
	if len(*args.Command) != 0 {
		fmt.Print(eng.RenderTooltip(*args.Command))
		return
	}
	if *args.RPrompt {
		fmt.Print(eng.RenderRPrompt())
		return
	}
	prompt := eng.Render()
	if !*args.ExportPNG {
		fmt.Print(prompt)
		return
	}
	imageCreator := &engine.ImageRenderer{
		AnsiString:    prompt,
		Author:        *args.Author,
		CursorPadding: *args.CursorPadding,
		RPromptOffset: *args.RPromptOffset,
		BgColor:       *args.BGColor,
		Ansi:          ansi,
	}
	imageCreator.Init(*args.Config)
	err := imageCreator.SavePNG()
	if err != nil {
		fmt.Print(err.Error())
	}
}
