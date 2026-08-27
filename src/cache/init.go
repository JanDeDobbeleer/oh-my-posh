package cache

import (
	"fmt"
	"os"
	"sync"
	"time"
	"uuid"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Option func()

var (
	sessionID  string
	newSession bool
	persist    bool
	noSession  bool
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

var NoSession Option = func() {
	log.Debug("disable session cache")
	noSession = true
}

func Init(shell string, options ...Option) {
	for _, opt := range options {
		opt()
	}

	Device.init(DeviceStore, persist)

	if noSession {
		return
	}

	sessionFileName := fmt.Sprintf("%s.%s.%s", shell, SessionID(), DeviceStore)
	Session.init(sessionFileName, persist)
}

func SessionID() string {
	defer log.Trace(time.Now())

	once.Do(func() {
		if newSession {
			sessionID = newSessionID()
			return
		}

		sessionID = os.Getenv("POSH_SESSION_ID")
		if sessionID == "" {
			sessionID = newSessionID()
		}
	})

	return sessionID
}

func Close() {
	Session.close()
	Device.close()
}

func newSessionID() string {
	return uuid.NewV4().String()
}
