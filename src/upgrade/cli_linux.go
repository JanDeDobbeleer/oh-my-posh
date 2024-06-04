//go:build !windows && !darwin

package upgrade

import "errors"

var successMsg string

func install() error {
	return errors.New("upgrade is not supported on this platform")
}
