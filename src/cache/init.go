package cache

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Option func()

var (
	sessionID  string
	newSession bool
	persist    bool
	once       sync.Once
)

var NewSession Option = func() {
	log.Debug("starting a new session")
	newSession = true
}

var Persist Option = func() {
	log.Debug("enable persistent cache")
	persist = true
}

func Init(shell string, options ...Option) {
	for _, opt := range options {
		opt()
	}

	sessionFileName := fmt.Sprintf("%s.%s.%s", shell, SessionID(), DeviceStore)
	Session.init(sessionFileName, persist)
	Device.init(DeviceStore, persist)
}

func SessionID() string {
	defer log.Trace(time.Now())

	once.Do(func() {
		if newSession {
			sessionID = uuid.NewString()
			return
		}

		sessionID = os.Getenv("POSH_SESSION_ID")
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
	})

	return sessionID
}

func Close() {
	Session.close()
	Device.close()
}
