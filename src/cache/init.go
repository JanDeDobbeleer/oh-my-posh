package cache

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

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

// newSessionID returns a random RFC 4122 version 4 UUID string, the same
// format github.com/google/uuid produced for session identifiers.
func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand never fails on supported platforms; fall back to a
		// time-derived id rather than panicking in the prompt path
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
