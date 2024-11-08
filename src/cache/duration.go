package cache

import (
	"time"
)

type Duration string

const (
	INFINITE = Duration("infinite")
	NONE     = Duration("none")
	ONEWEEK  = Duration("168h")
	ONEDAY   = Duration("24h")
	TWOYEARS = Duration("17520h")
)

func (d Duration) Seconds() int {
	if d == NONE {
		return 0
	}

	if d == INFINITE {
		return -1
	}

	duration, err := time.ParseDuration(string(d))
	if err != nil {
		return 0
	}

	return int(duration.Seconds())
}

func (d Duration) IsEmpty() bool {
	return len(d) == 0
}

func ToDuration(seconds int) Duration {
	if seconds == 0 {
		return ""
	}

	if seconds == -1 {
		return INFINITE
	}

	duration := time.Duration(seconds) * time.Second
	return Duration(duration.String())
}
