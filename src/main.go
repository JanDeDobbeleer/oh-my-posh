package main

import "github.com/jandedobbeleer/oh-my-posh/src/cli"

var (
	Version = "development"
)

func main() {
	cli.Execute(Version)
}
