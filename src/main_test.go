package main

import (
	"bytes"
	"oh-my-posh/cli"
	"testing"
)

func BenchmarkInit(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"init", "fish", "--print", "--config", "err.omp.json"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}

func BenchmarkPrimary(b *testing.B) {
	cmd := cli.RootCmd
	// needs to be a non-existing file as we panic otherwise
	cmd.SetArgs([]string{"print", "primary", "--config", "err.omp.json", "--pwd", "/Users/jan/Code/oh-my-posh/src"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}
