//go:build !darwin

package environment

import (
	"oh-my-posh/environment/battery"
	"time"
)

func (env *ShellEnvironment) BatteryState() (*battery.Info, error) {
	defer env.Trace(time.Now(), "BatteryState")
	info, err := battery.Get()
	if err != nil {
		env.Log(Error, "BatteryState", err.Error())
		return nil, err
	}
	return info, nil
}
