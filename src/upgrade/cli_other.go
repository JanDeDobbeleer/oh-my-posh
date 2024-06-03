//go:build linux || freebsd || openbsd || netbsd

package upgrade

import "github.com/jandedobbeleer/oh-my-posh/src/platform"

func Run(_ platform.Environment) {}

var (
	Supported = false
)
