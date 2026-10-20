package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

type exitCodeError int

func (e exitCodeError) Error() string {
	return "exit status"
}

func (e exitCodeError) ExitCode() int {
	return int(e)
}

func TestSSHAgent(t *testing.T) {
	const twoKeys = "256 SHA256:abc /home/user/.ssh/id_ed25519 (ED25519)\n" +
		"3072 SHA256:def work@laptop (RSA)"

	cases := []struct {
		Error           error
		Case            string
		GOOS            string
		AuthSock        string
		Output          string
		ExpectedCount   int
		ExpectedRunning bool
		SkipsCommand    bool
	}{
		{
			Case:         "no socket on linux",
			GOOS:         runtime.LINUX,
			SkipsCommand: true,
		},
		{
			Case:            "keys loaded",
			GOOS:            runtime.LINUX,
			AuthSock:        "/tmp/agent.sock",
			Output:          twoKeys,
			ExpectedRunning: true,
			ExpectedCount:   2,
		},
		{
			Case:            "agent without identities",
			GOOS:            runtime.DARWIN,
			AuthSock:        "/tmp/agent.sock",
			Output:          "The agent has no identities.",
			Error:           exitCodeError(1),
			ExpectedRunning: true,
		},
		{
			Case:     "agent unreachable",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/stale.sock",
			Output:   "Could not open a connection to your authentication agent.",
			Error:    exitCodeError(2),
		},
		{
			Case:     "ssh-add missing",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/agent.sock",
			Error:    errors.New("executable file not found in $PATH"),
		},
		{
			Case:            "windows named pipe without socket variable",
			GOOS:            runtime.WINDOWS,
			Output:          twoKeys,
			ExpectedRunning: true,
			ExpectedCount:   2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("GOOS").Return(tc.GOOS)
			env.On("Getenv", "SSH_AUTH_SOCK").Return(tc.AuthSock)
			env.On("RunCommand", "ssh-add", []string{"-l"}).Return(tc.Output, tc.Error)

			agent := &SSHAgent{}
			agent.Init(options.Map{}, env)

			assert.True(t, agent.Enabled())
			assert.Equal(t, tc.ExpectedRunning, agent.Running)
			assert.Equal(t, tc.ExpectedCount, agent.Count)

			if tc.SkipsCommand {
				env.AssertNotCalled(t, "RunCommand", "ssh-add", []string{"-l"})
			}
		})
	}
}
