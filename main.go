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
	engine := &engine{
		settings: settings,
		env:      env,
		renderer: &ColorWriter{
			Buffer: new(bytes.Buffer),
		},
	}
	prompt := engine.string()
	fmt.Print(prompt)
}
