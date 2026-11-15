package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

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
	const fiveKeys = twoKeys + "\n" +
		"256 SHA256:ghi build@ci (ED25519)\n" +
		"256 SHA256:jkl /home/user/.ssh/id_work (ED25519)\n" +
		"256 SHA256:mno deploy@host (ED25519)"

	cases := []struct {
		Error                 error
		Options               options.Map
		Case                  string
		ExpectedKeys          template.Markup
		GOOS                  string
		AuthSock              string
		Output                string
		ExpectedCount         int
		ExpectedUnmappedCount int
		ExpectedRunning       bool
		SkipsCommand          bool
	}{
		{
			Case:         "no socket on linux",
			GOOS:         runtime.LINUX,
			SkipsCommand: true,
		},
		{
			Case:                  "keys loaded",
			GOOS:                  runtime.LINUX,
			AuthSock:              "/tmp/agent.sock",
			Output:                twoKeys,
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedKeys:          "/home/user/.ssh/id_ed25519 work@laptop",
			ExpectedUnmappedCount: 2,
		},
		{
			Case:     "mapped key replaces comment",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/agent.sock",
			Output:   twoKeys,
			Options: options.Map{
				SSHAgentMappedKeys: map[string]any{"SHA256:abc": "<green>personal</>"},
			},
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedKeys:          "<green>personal</> work@laptop",
			ExpectedUnmappedCount: 1,
		},
		{
			Case:     "key template and separator",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/agent.sock",
			Output:   twoKeys,
			Options: options.Map{
				SSHAgentKeyTemplate:  "{{ .Type }}/{{ .Bits }}/{{ .Fingerprint }}",
				SSHAgentKeySeparator: " | ",
			},
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedKeys:          "ED25519/256/SHA256:abc | RSA/3072/SHA256:def",
			ExpectedUnmappedCount: 2,
		},
		{
			Case:     "unmapped keys are counted and omitted without extra separators",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/agent.sock",
			Output:   fiveKeys,
			Options: options.Map{
				SSHAgentKeyTemplate:  "{{ .Alias }}",
				SSHAgentKeySeparator: ", ",
				SSHAgentMappedKeys: map[string]any{
					"SHA256:def": "work",
					"SHA256:jkl": "personal",
				},
			},
			ExpectedRunning:       true,
			ExpectedCount:         5,
			ExpectedUnmappedCount: 3,
			ExpectedKeys:          "work, personal",
		},
		{
			Case:     "empty key templates do not render separators",
			GOOS:     runtime.LINUX,
			AuthSock: "/tmp/agent.sock",
			Output:   twoKeys,
			Options: options.Map{
				SSHAgentKeyTemplate:  "{{ .Alias }}",
				SSHAgentKeySeparator: ", ",
			},
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedUnmappedCount: 2,
		},
		{
			Case:                  "comment with spaces and chevrons stays data",
			GOOS:                  runtime.LINUX,
			AuthSock:              "/tmp/agent.sock",
			Output:                "256 SHA256:abc my <red>key</> (ED25519)",
			ExpectedRunning:       true,
			ExpectedCount:         1,
			ExpectedKeys:          template.EscapeMarkup("my <red>key</>"),
			ExpectedUnmappedCount: 1,
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
			Case:                  "windows named pipe without socket variable",
			GOOS:                  runtime.WINDOWS,
			Output:                twoKeys,
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedKeys:          "/home/user/.ssh/id_ed25519 work@laptop",
			ExpectedUnmappedCount: 2,
		},
		{
			Case:                  "windows line endings",
			GOOS:                  runtime.WINDOWS,
			Output:                strings.ReplaceAll(twoKeys, "\n", "\r\n"),
			Options:               options.Map{SSHAgentKeyTemplate: "{{ .Comment }}:{{ .Type }}"},
			ExpectedRunning:       true,
			ExpectedCount:         2,
			ExpectedKeys:          "/home/user/.ssh/id_ed25519:ED25519 work@laptop:RSA",
			ExpectedUnmappedCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("GOOS").Return(tc.GOOS)
			env.On("Getenv", "SSH_AUTH_SOCK").Return(tc.AuthSock)
			env.On("RunCommand", "ssh-add", []string{"-l"}).Return(tc.Output, tc.Error)
			env.On("Shell").Return("bash")

			origCache := template.Cache
			t.Cleanup(func() { template.Cache = origCache })
			template.Cache = &cache.Template{}
			template.Init(env, nil, nil)

			agent := &SSHAgent{}
			agent.Init(tc.Options, env)

			assert.True(t, agent.Enabled())
			assert.Equal(t, tc.ExpectedRunning, agent.Running)
			assert.Equal(t, tc.ExpectedCount, agent.Count)
			assert.Equal(t, tc.ExpectedUnmappedCount, agent.UnmappedCount)
			assert.Equal(t, tc.ExpectedKeys, agent.Keys)

			if tc.SkipsCommand {
				env.AssertNotCalled(t, "RunCommand", "ssh-add", []string{"-l"})
			}
		})
	}
}
