// Package tui runs the interactive device-code authentication flow for `oh-my-posh auth`: it
// walks GitHub Copilot's and the YouTube Music Desktop App's device-code exchange through to a
// stored token, showing which step it is on.
//
// It lives one level below cli/auth because cli/auth holds only the plain cache-key constants
// (CopilotTokenKey, YTMDABASEURL, YTMDATOKEN) that the segments package needs to read a token back
// out of the cache, and that import graph also has to compile for wasm - a target with no terminal
// to authenticate against. Keeping the device flow here means importing cli/auth for its constants
// never drags a terminal UI along; this package imports cli/auth, never the other way around.
package tui

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type state int

const (
	code state = iota
	token
	done
)

// status is the line the running flow paints on. Package-level because the flows report progress
// from deep inside their own polling loops, the same way they did when this drove a program.
var status *ui.Status

// current is the flow being run, so a state change can ask it what to say. The flows word their
// own steps - GitHub names the device code, YouTube Music names its base URL - so the message is
// theirs to render, not this file's.
var current Flow

func setState(next state) {
	if current == nil || status == nil {
		return
	}

	current.base().state = next

	if next == done {
		return
	}

	status.Set(current.message())
}

type model struct {
	env    runtime.Environment
	err    error
	status func(error) string
	code   string
	state  state
}

// base lets Run reach the embedded model of whichever flow it was handed. Unexported, so the
// interface is closed to this package's own flows.
func (m *model) base() *model {
	return m
}

func (m *model) message() string {
	switch m.state {
	case code:
		return "Fetching code for authentication"
	case token:
		return fmt.Sprintf("Fetching token with code: %s", m.code)
	default:
		return ""
	}
}

// Flow is one provider's device-code exchange. Authenticate runs it to completion, reporting
// progress through setState and leaving its outcome on the embedded model.
type Flow interface {
	Authenticate()
	base() *model
	message() string
}

func Run(flow Flow) error {
	current = flow

	status = ui.NewStatus(os.Stdout)
	status.Start(flow.message())

	defer func() {
		current = nil
		status = nil
	}()

	// Synchronous: this used to run in a goroutine so a message loop could keep drawing, and the
	// status line now draws itself from its own ticker.
	flow.Authenticate()

	base := flow.base()

	status.Stop(base.status(base.err))

	return base.err
}
