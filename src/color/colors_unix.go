//go:build !windows && !darwin

package color

import "github.com/jandedobbeleer/oh-my-posh/src/runtime"

func GetAccentColor(_ runtime.Environment) (*RGB, error) {
	return nil, &runtime.NotImplemented{}
}
