//go:build windows && !386

package main

const (
	dotnetExitCode = int(0x80008091)
)
