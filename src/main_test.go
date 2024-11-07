package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cli"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
)

func BenchmarkInit(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"init", "fish", "--print", "--silent"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}

func BenchmarkPrimary(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"print", prompt.PRIMARY, "--pwd", "/Users/jan/Code/oh-my-posh/src", "--shell", "fish", "--silent"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}

	fmt.Println("")
}
