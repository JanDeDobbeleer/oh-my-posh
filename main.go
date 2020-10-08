package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
)

type args struct {
	ErrorCode   *int
	PrintConfig *bool
	Config      *string
	Shell       *string
	PWD         *string
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
			"Config prints the current settings in json format"),
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
	}
	flag.Parse()
	env := &environment{
		args: args,
	}
	settings := GetSettings(env)
	if *args.PrintConfig {
		theme, _ := json.MarshalIndent(settings, "", "    ")
		fmt.Println(string(theme))
		return
	}
	colorWriter := &Renderer{
		Buffer: new(bytes.Buffer),
	}
	shell := env.getShellName()
	if *args.Shell != "" {
		shell = *args.Shell
	}
	colorWriter.init(shell)
	engine := &engine{
		settings: settings,
		env:      env,
		renderer: colorWriter,
	}
	engine.render()
}
