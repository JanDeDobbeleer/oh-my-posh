//go:build !darwin

package platform

import (
	"oh-my-posh/platform/battery"
	"time"
)

func (env *Shell) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now(), "BatteryState")
	info, err := battery.Get()
	if err != nil {
		env.Log(Error, "BatteryState", err.Error())
		return nil, err
	}
	return info, nil
}
