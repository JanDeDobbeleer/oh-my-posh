package path

import (
	"runtime"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const (
	windows = "windows"
)

func Separator() string {
	defer log.Trace(time.Now())

	if runtime.GOOS == windows {
		return `\`
	}

	return "/"
}

func IsSeparator(c uint8) bool {
	if c == '/' {
		return true
	}

	if runtime.GOOS == windows && c == '\\' {
		return true
	}

	return false
}
