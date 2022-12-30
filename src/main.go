package main

import "github.com/jandedobbeleer/oh-my-posh/cli"

var (
	Version = "development"
)

func main() {
	cli.Execute(Version)
}
