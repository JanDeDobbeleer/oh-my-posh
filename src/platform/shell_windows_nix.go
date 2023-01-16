//go:build !darwin

package platform

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/platform/battery"
)

func (env *Shell) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now())
	info, err := battery.Get()
	if err != nil {
		env.Error(err)
		return nil, err
	}
	return info, nil
}
