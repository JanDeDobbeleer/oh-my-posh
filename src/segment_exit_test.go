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
			env: env,
		}
		assert.Equal(t, tc.Expected, e.enabled())
	}
}

func TestExitWriterFormattedText(t *testing.T) {
	cases := []struct {
		ExitCode        int
		Expected        string
		SuccessIcon     string
		ErrorIcon       string
		DisplayExitCode bool
		AlwaysNumeric   bool
	}{
		{ExitCode: 129, Expected: "SIGHUP", DisplayExitCode: true},
		{ExitCode: 5001, Expected: "5001", DisplayExitCode: true},
		{ExitCode: 147, Expected: "SIGSTOP", DisplayExitCode: true},
		{ExitCode: 147, Expected: "", DisplayExitCode: false},
		{ExitCode: 147, Expected: "147", DisplayExitCode: true, AlwaysNumeric: true},
		{ExitCode: 0, Expected: "wooopie", SuccessIcon: "wooopie"},
		{ExitCode: 129, Expected: "err SIGHUP", ErrorIcon: "err ", DisplayExitCode: true},
		{ExitCode: 129, Expected: "err", ErrorIcon: "err", DisplayExitCode: false},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(tc.ExitCode)
		props := &properties{
			foreground: "#111111",
			background: "#ffffff",
			values: map[Property]interface{}{
				SuccessIcon:     tc.SuccessIcon,
				ErrorIcon:       tc.ErrorIcon,
				DisplayExitCode: tc.DisplayExitCode,
				AlwaysNumeric:   tc.AlwaysNumeric,
			},
		}
		e := &exit{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.Expected, e.getFormattedText())
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
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(exitcode)
		e := &exit{
			env: env,
		}
		assert.Equal(t, want, e.getMeaningFromExitCode())
	}
}

func TestAlwaysNumericExitCode(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("lastErrorCode", nil).Return(1)
	props := &properties{
		values: map[Property]interface{}{
			AlwaysNumeric: true,
		},
	}
	e := &exit{
		env:   env,
		props: props,
	}
	assert.Equal(t, "1", e.getMeaningFromExitCode())
}
