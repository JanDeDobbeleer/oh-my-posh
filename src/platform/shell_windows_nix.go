//go:build !darwin

package platform

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/platform/battery"
)

func (env *Shell) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now(), "BatteryState")
	info, err := battery.Get()
	if err != nil {
		env.Error("BatteryState", err)
		return nil, err
	}
	return info, nil
}
