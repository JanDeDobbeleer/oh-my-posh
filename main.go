package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

// Version number of oh-my-posh
var Version = "development"

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
	}
	flag.Parse()
	env := &environment{
		args: args,
	}
	if *args.Millis {
		fmt.Print(time.Now().UnixNano() / 1000000)
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
	shell := env.getShellName()
	if *args.Shell != "" {
		shell = *args.Shell
	}
	renderer := &AnsiRenderer{
		buffer: new(bytes.Buffer),
	}
	colorer := &AnsiColor{
		buffer: new(bytes.Buffer),
	}
	renderer.init(shell)
	colorer.init(shell)
	engine := &engine{
		settings: settings,
		env:      env,
		color:    colorer,
		renderer: renderer,
	}
	engine.render()
}
