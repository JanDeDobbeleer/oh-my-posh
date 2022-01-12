package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitWriterEnabled(t *testing.T) {
	cases := []struct {
		ExitCode int
		Expected bool
	}{
		{ExitCode: 102, Expected: true},
		{ExitCode: 0, Expected: false},
		{ExitCode: -1, Expected: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(tc.ExitCode)
		e := &exit{
			env:   env,
			props: properties{},
		}
		assert.Equal(t, tc.Expected, e.enabled())
	}
}

func TestGetMeaningFromExitCode(t *testing.T) {
	errorMap := make(map[int]string)
	errorMap[1] = "ERROR"
	errorMap[2] = "USAGE"
	errorMap[126] = "NOPERM"
	errorMap[127] = "NOTFOUND"
	errorMap[129] = "SIGHUP"
	errorMap[130] = "SIGINT"
	errorMap[131] = "SIGQUIT"
	errorMap[132] = "SIGILL"
	errorMap[133] = "SIGTRAP"
	errorMap[134] = "SIGIOT"
	errorMap[135] = "SIGBUS"
	errorMap[136] = "SIGFPE"
	errorMap[137] = "SIGKILL"
	errorMap[138] = "SIGUSR1"
	errorMap[139] = "SIGSEGV"
	errorMap[140] = "SIGUSR2"
	errorMap[141] = "SIGPIPE"
	errorMap[142] = "SIGALRM"
	errorMap[143] = "SIGTERM"
	errorMap[144] = "SIGSTKFLT"
	errorMap[145] = "SIGCHLD"
	errorMap[146] = "SIGCONT"
	errorMap[147] = "SIGSTOP"
	errorMap[148] = "SIGTSTP"
	errorMap[149] = "SIGTTIN"
	errorMap[150] = "SIGTTOU"
	errorMap[151] = "151"
	errorMap[7000] = "7000"
	for exitcode, want := range errorMap {
		e := &exit{}
		e.code = exitcode
		assert.Equal(t, want, e.getMeaningFromExitCode())
	}
}

func TestExitWriterTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		ExitCode int
		Expected string
		Template string
	}{
		{Case: "Only code", ExitCode: 129, Expected: "129", Template: "{{ .Code }}"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(tc.ExitCode)
		env.onTemplate()
		props := properties{
			SegmentTemplate: tc.Template,
		}
		e := &exit{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.Expected, e.string(), tc.Case)
	}
}
