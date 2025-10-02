package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	enabled bool
	raw     bool

	log strings.Builder
)

func Enable(plain bool) {
	enabled = true
	raw = plain

	Debugf("logging enabled, raw mode: %t", plain)
}

func Trace(start time.Time, args ...string) {
	if !enabled {
		return
	}

	elapsed := time.Since(start)
	fn, _ := funcSpec()

	// Color-code elapsed time based on duration
	var coloredElapsed Text
	ms := elapsed.Milliseconds()

	switch {
	case ms < 1:
		coloredElapsed = Text(elapsed.String()).Green().Plain()
	case ms >= 1 && ms < 10:
		coloredElapsed = Text(elapsed.String()).Yellow().Plain()
	case ms >= 10 && ms < 100:
		coloredElapsed = Text(elapsed.String()).Orange().Plain()
	default: // >= 100ms
		coloredElapsed = Text(elapsed.String()).Red().Plain()
	}

	header := fmt.Sprintf("%s(%s) - %s", fn, strings.Join(args, " "), coloredElapsed)

	printLn(trace, header)
}

func Debug(message ...string) {
	if !enabled {
		return
	}

	fn, line := funcSpec()
	header := fmt.Sprintf("%s:%d", fn, line)

	printLn(debug, header, strings.Join(message, " "))
}

func Debugf(format string, args ...any) {
	if !enabled {
		return
	}

	message := fmt.Sprintf(format, args...)
	Debug(message)
}

func Error(err error) {
	if !enabled {
		return
	}
	fn, line := funcSpec()
	header := fmt.Sprintf("%s:%d", fn, line)

	printLn(bug, header, err.Error())
}

func Errorf(format string, args ...any) {
	if !enabled {
		return
	}

	Error(fmt.Errorf(format, args...))
}

func String() string {
	return log.String()
}

func funcSpec() (string, int) {
	pcs := make([]uintptr, 4)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return "", 0
	}

	frames := runtime.CallersFrames(pcs[:n])
	var frame runtime.Frame
	more := true

	// Loop through frames until we're out of log.go
	for more {
		frame, more = frames.Next()
		if strings.Contains(frame.File, "log.go") {
			continue
		}

		// Found first non-log.go frame
		fn := frame.Function
		fn = fn[strings.LastIndex(fn, ".")+1:]
		file := filepath.Base(frame.File)

		if strings.HasPrefix(fn, "func") {
			return file, frame.Line
		}

		return fmt.Sprintf("%s:%s", file, fn), frame.Line
	}

	return "", 0
}
