package main

import (
	"bytes"
	"oh-my-posh/cli"
	"testing"
)

func BenchmarkInit(b *testing.B) {
	cmd := cli.RootCmd
	cmd.SetArgs([]string{"init", "fish", "--print", "--config", "err.omp.json"})
	out := bytes.NewBufferString("")
	cmd.SetOut(out)

	for i := 0; i < b.N; i++ {
		_ = cmd.Execute()
	}
}
