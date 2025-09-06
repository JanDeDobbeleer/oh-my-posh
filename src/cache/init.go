package cache

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func SessionID() string {
	defer log.Trace(time.Now())

	once.Do(func() {
		sessionID = os.Getenv("POSH_SESSION_ID")
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
	})

	return sessionID
}

func Init(shell string, persist bool) {
	sessionFileName := fmt.Sprintf("%s.%s.%s", shell, SessionID(), FileName)
	Session.init(sessionFileName, persist)
	Device.init(FileName, persist)
}

func Close() {
	Session.close()
	Device.close()
}
