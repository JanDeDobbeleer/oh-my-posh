package main

import "oh-my-posh/cli"

var (
	Version = "development"
)

func main() {
	cli.Execute(Version)
}
