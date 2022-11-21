package log

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var enabled bool
var logBuilder strings.Builder

func Enable() {
	enabled = true
	log.SetOutput(&logBuilder)
}

func Info(message string) {
	if !enabled {
		return
	}
	log.Println(message)
}

func Trace(start time.Time, function string, args ...string) {
	if !enabled {
		return
	}
	elapsed := time.Since(start)
	var argString string
	if len(args) > 0 {
		argString = fmt.Sprintf(", args: %s", strings.Join(args, " "))
	}
	trace := entry(fmt.Sprintf("%s: %s%s", function, elapsed, argString))
	log.Println(trace)
}

func Debug(funcName, message string) {
	if !enabled {
		return
	}
	trace := entry(fmt.Sprintf("%s\n%s", funcName, message))
	trace.Format(Yellow)
	log.Println(trace)
}

func DebugF(function string, fn func() string) {
	if !enabled {
		return
	}
	trace := entry(fmt.Sprintf("%s\n%s", function, fn()))
	trace.Format(Yellow)
	log.Println(trace)
}

func Error(funcName string, err error) {
	if !enabled {
		return
	}
	trace := entry(fmt.Sprintf("%s\n%s", funcName, err.Error()))
	trace.Format(Red)
	log.Println(trace)
}

func String() string {
	return logBuilder.String()
}
