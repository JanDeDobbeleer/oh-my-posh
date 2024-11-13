//go:build !darwin

package runtime

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
)

func (term *Terminal) BatteryState() (*battery.Info, error) {
	defer log.Trace(time.Now())
	info, err := battery.Get()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return info, nil
}
