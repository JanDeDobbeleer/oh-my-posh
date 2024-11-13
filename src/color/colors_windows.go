package color

import (
	"errors"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

func GetAccentColor(env runtime.Environment) (*RGB, error) {
	defer log.Trace(time.Now())

	if env == nil {
		return nil, errors.New("unable to get color without environment")
	}

	// see https://stackoverflow.com/questions/3560890/vista-7-how-to-get-glass-color
	value, err := env.WindowsRegistryKeyValue(`HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM\ColorizationColor`)
	if err != nil || value.ValueType != runtime.DWORD {
		return nil, err
	}

	return &RGB{
		R: byte(value.DWord >> 16),
		G: byte(value.DWord >> 8),
		B: byte(value.DWord),
	}, nil
}
