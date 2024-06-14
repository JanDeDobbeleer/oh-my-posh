//go:build linux || freebsd || openbsd || netbsd

package upgrade

import "github.com/jandedobbeleer/oh-my-posh/src/platform"

const (
	upgradeNotice = `
A new release of Oh My Posh is available: %s → %s

To upgrade, use your favorite package manager or, if you used Homebrew to install, run: 'brew update; brew upgrade oh-my-posh'
`
)

func Run(_ platform.Environment) {}

const (
	Supported = false
)
