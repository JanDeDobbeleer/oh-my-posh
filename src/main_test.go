package main

import (
	"bytes"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cli"
)

func BenchmarkInit(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"init", "fish", "--print"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}

func BenchmarkPrimary(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"print", "primary", "--pwd", "/Users/jan/Code/oh-my-posh/src", "--shell", "fish"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}
