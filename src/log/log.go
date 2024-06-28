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
	plain   bool
	log     strings.Builder
)

func Enable() {
	enabled = true
}

func Plain() {
	plain = true
}

func Trace(start time.Time, args ...string) {
	if !enabled {
		return
	}

	elapsed := time.Since(start)
	fn, _ := funcSpec()
	header := fmt.Sprintf("%s(%s) - %s", fn, strings.Join(args, " "), Text(elapsed.String()).Yellow().Plain())

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
