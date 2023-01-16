package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var enabled bool
var log strings.Builder

func Enable() {
	enabled = true
}

func Info(message string) {
	if !enabled {
		return
	}
	log.WriteString(message)
}

func Trace(start time.Time, args ...string) {
	if !enabled {
		return
	}
	elapsed := time.Since(start)
	fn, _ := funcSpec()
	header := fmt.Sprintf("%s(%s) - \x1b[38;2;156;231;201m%s\033[0m", fn, strings.Join(args, " "), elapsed)
	printLn(trace, header)
}

func Debug(message string) {
	if !enabled {
		return
	}
	fn, line := funcSpec()
	header := fmt.Sprintf("%s:%d", fn, line)
	printLn(debug, header, message)
}

func DebugF(fn func() string) {
	if !enabled {
		return
	}
	fn2, line := funcSpec()
	header := fmt.Sprintf("%s:%d", fn2, line)
	printLn(debug, header, fn())
}

func Error(err error) {
	if !enabled {
		return
	}
	fn, line := funcSpec()
	header := fmt.Sprintf("%s:%d", fn, line)
	printLn(bug, header, err.Error())
}

func String() string {
	return log.String()
}

func funcSpec() (string, int) {
	pc, file, line, OK := runtime.Caller(3)
	if !OK {
		return "", 0
	}
	fn := runtime.FuncForPC(pc).Name()
	fn = fn[strings.LastIndex(fn, ".")+1:]
	file = filepath.Base(file)
	if strings.HasPrefix(fn, "func") {
		return file, line
	}
	return fmt.Sprintf("%s:%s", file, fn), line
}
