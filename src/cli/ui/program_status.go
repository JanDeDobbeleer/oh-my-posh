package ui

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	programStatusApp              = "oh-my-posh"
	programStatusMessageByteLimit = 2048
)

// ProgramStatus reports foreground CLI work through OSC 7501.
type ProgramStatus struct {
	writer io.Writer
	last   string
	mutex  sync.Mutex
}

func NewProgramStatus(writer io.Writer) *ProgramStatus {
	return &ProgramStatus{writer: writer}
}

func (s *ProgramStatus) Working(message string) {
	s.write("working", message, nil)
}

func (s *ProgramStatus) Progress(message string, percentage int) {
	percentage = min(max(percentage, 0), 100)
	s.write("working", message, &percentage)
}

func (s *ProgramStatus) Blocked(kind, message string) {
	s.write("blocked", message, nil, "kind="+kind)
}

func (s *ProgramStatus) Done(message string) {
	s.write("done", message, nil)
}

func (s *ProgramStatus) Error(message string) {
	s.write("error", message, nil)
}

func (s *ProgramStatus) write(state, message string, progress *int, extra ...string) {
	pairs := []string{
		"state=" + state,
		"app=" + programStatusApp,
	}

	pairs = append(pairs, extra...)

	if progress != nil {
		pairs = append(pairs, fmt.Sprintf("progress=%d", *progress))
	}

	if message = programStatusMessage(message); message != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(message))
		pairs = append(pairs, "msg="+encoded)
	}

	report := "\x1b]7501;" + strings.Join(pairs, ":") + "\x1b\\"

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if report == s.last {
		return
	}

	s.last = report
	_, _ = fmt.Fprint(s.writer, report)
}

func programStatusMessage(message string) string {
	message = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, message)

	if len(message) <= programStatusMessageByteLimit {
		return message
	}

	message = message[:programStatusMessageByteLimit]
	for !utf8.ValidString(message) {
		message = message[:len(message)-1]
	}

	return message
}
