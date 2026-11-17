package segments

import (
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

var (
	gpgKeyInfoArgs = []string{"--no-autostart", "keyinfo --list", "/bye"}
	gpgInfoArgs    = []string{"--no-autostart", "GETINFO pid", "/bye"}
)

const (
	gpgGripA = "A2E8B723CE2E4B367F0B5C3ADCF1F1C1B5698B08"
	gpgGripB = "96C783332F2A5FC8705A983186E327BF405BD428"
	gpgGripC = "665D86B8E1D5140BB429E9EC22E43CC39EE3123D"

	gpgAgentInfo = "D 4242\nOK"
)

type gpgExitError int

func (e gpgExitError) Error() string {
	return "exit status"
}

func (e gpgExitError) ExitCode() int {
	return int(e)
}

func gpgKeyInfo(lines ...string) string {
	return strings.Join(append(lines, "OK"), "\n")
}

func TestGPGAgent(t *testing.T) {
	cachedA := "S KEYINFO " + gpgGripA + " D - - 1 P - 600 -"
	lockedB := "S KEYINFO " + gpgGripB + " D - - - P - - -"
	cachedC := "S KEYINFO " + gpgGripC + " D - - 1 P - 90 -"

	cases := []struct {
		AgentError          error
		GitError            error
		Options             options.Map
		Case                string
		AgentOutput         string
		InfoOutput          string
		GitOutput           string
		GPGTTY              string
		AuthSock            string
		ExpectedKeys        template.Markup
		ExpectedSigningKey  string
		ExpectedCount       int
		ExpectedCachedCount int
		FetchKeys           bool
		MissingGPG          bool
		MissingGit          bool
		ExpectedRunning     bool
		ExpectedSignCommits bool
		ExpectedTTYSet      bool
		ExpectedSSHSupport  bool
	}{
		{
			Case:            "the agent is detected without reading the passphrase cache",
			InfoOutput:      gpgAgentInfo,
			ExpectedRunning: true,
		},
		{
			Case: "no agent answers GETINFO",
		},
		{
			Case:            "GETINFO with windows line endings",
			InfoOutput:      strings.ReplaceAll(gpgAgentInfo, "\n", "\r\n"),
			ExpectedRunning: true,
		},
		{
			Case:       "gpg-connect-agent missing",
			MissingGPG: true,
		},
		{
			Case:                "only unlocked keys are listed by default",
			FetchKeys:           true,
			AgentOutput:         gpgKeyInfo(cachedA, lockedB, cachedC),
			ExpectedRunning:     true,
			ExpectedCount:       3,
			ExpectedCachedCount: 2,
			ExpectedKeys:        template.RawMarkup(gpgGripA + " " + gpgGripC),
		},
		{
			Case:            "agent running without keys",
			FetchKeys:       true,
			AgentOutput:     gpgKeyInfo(),
			ExpectedRunning: true,
		},
		{
			Case:      "agent not running reports an error on stderr but exits 0",
			FetchKeys: true,
		},
		{
			Case:       "gpg-connect-agent missing while fetching keys",
			FetchKeys:  true,
			MissingGPG: true,
		},
		{
			Case:       "gpg-connect-agent fails",
			FetchKeys:  true,
			AgentError: errors.New("exit status 2"),
		},
		{
			Case:        "mapped keygrips use their alias",
			FetchKeys:   true,
			AgentOutput: gpgKeyInfo(cachedA, cachedC),
			Options: options.Map{
				GPGAgentMappedKeys:   map[string]any{gpgGripA: "<green>work</>"},
				GPGAgentKeySeparator: ", ",
			},
			ExpectedRunning:     true,
			ExpectedCount:       2,
			ExpectedCachedCount: 2,
			ExpectedKeys:        template.RawMarkup("<green>work</>, " + gpgGripC),
		},
		{
			Case:        "key template exposes ttl and locked keys",
			FetchKeys:   true,
			AgentOutput: gpgKeyInfo(cachedA, lockedB),
			Options: options.Map{
				GPGAgentKeyTemplate:  "{{ if .Cached }}{{ .TTL }}s{{ else }}locked{{ end }}",
				GPGAgentKeySeparator: " | ",
			},
			ExpectedRunning:     true,
			ExpectedCount:       2,
			ExpectedCachedCount: 1,
			ExpectedKeys:        "600s | locked",
		},
		{
			Case:        "keyinfo without the ttl and flags fields",
			FetchKeys:   true,
			AgentOutput: gpgKeyInfo("S KEYINFO " + gpgGripA + " D - - 1 P -"),
			Options: options.Map{
				GPGAgentKeyTemplate: "{{ .Keygrip }}:{{ .TTL }}",
			},
			ExpectedRunning:     true,
			ExpectedCount:       1,
			ExpectedCachedCount: 1,
			ExpectedKeys:        template.RawMarkup(gpgGripA + ":0"),
		},
		{
			Case:            "keyinfo without the cached field",
			FetchKeys:       true,
			AgentOutput:     gpgKeyInfo("S KEYINFO " + gpgGripA + " D - -"),
			ExpectedRunning: true,
			ExpectedCount:   1,
		},
		{
			Case:            "a truncated keyinfo line is ignored",
			FetchKeys:       true,
			AgentOutput:     gpgKeyInfo("S KEYINFO " + gpgGripA + " D"),
			ExpectedRunning: true,
		},
		{
			Case:        "windows line endings",
			FetchKeys:   true,
			AgentOutput: strings.ReplaceAll(gpgKeyInfo(cachedA), "\n", "\r\n"),
			Options: options.Map{
				GPGAgentKeyTemplate: "{{ .TTL }}",
			},
			ExpectedRunning:     true,
			ExpectedCount:       1,
			ExpectedCachedCount: 1,
			ExpectedKeys:        "600",
		},
		{
			Case:                "git signing configuration",
			InfoOutput:          gpgAgentInfo,
			GitOutput:           "user.signingkey ABCDEF\ncommit.gpgsign true",
			ExpectedRunning:     true,
			ExpectedSigningKey:  "ABCDEF",
			ExpectedSignCommits: true,
		},
		{
			Case:               "git signing key without gpgsign",
			InfoOutput:         gpgAgentInfo,
			GitOutput:          "user.signingkey ABCDEF",
			ExpectedRunning:    true,
			ExpectedSigningKey: "ABCDEF",
		},
		{
			Case:                "git boolean spellings",
			InfoOutput:          gpgAgentInfo,
			GitOutput:           "commit.gpgsign yes",
			ExpectedRunning:     true,
			ExpectedSignCommits: true,
		},
		{
			Case:            "git exiting 1 means nothing is configured",
			InfoOutput:      gpgAgentInfo,
			GitError:        gpgExitError(1),
			ExpectedRunning: true,
		},
		{
			Case:            "git missing",
			InfoOutput:      gpgAgentInfo,
			MissingGit:      true,
			ExpectedRunning: true,
		},
		{
			Case:               "tty and ssh support",
			InfoOutput:         gpgAgentInfo,
			GPGTTY:             "/dev/pts/1",
			AuthSock:           "/run/user/1000/gnupg/S.gpg-agent.ssh",
			ExpectedRunning:    true,
			ExpectedTTYSet:     true,
			ExpectedSSHSupport: true,
		},
		{
			Case:            "plain ssh-agent is not gpg ssh support",
			InfoOutput:      gpgAgentInfo,
			AuthSock:        "/tmp/ssh-XXXX/agent.123",
			ExpectedRunning: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			cache.Device.DeleteAll()
			cache.Session.DeleteAll()
			t.Cleanup(func() {
				cache.Device.DeleteAll()
				cache.Session.DeleteAll()
			})

			env := new(mock.Environment)
			env.On("Getenv", "GPG_TTY").Return(tc.GPGTTY)
			env.On("Getenv", "SSH_AUTH_SOCK").Return(tc.AuthSock)
			env.On("Pwd").Return("/home/user/project")
			env.On("Shell").Return("bash")
			env.On("HasCommand", "gpg-connect-agent").Return(!tc.MissingGPG)
			env.On("HasCommand", "git").Return(!tc.MissingGit)
			env.On("RunCommand", "gpg-connect-agent", gpgKeyInfoArgs).
				Return(tc.AgentOutput, tc.AgentError)
			env.On("RunCommand", "gpg-connect-agent", gpgInfoArgs).
				Return(tc.InfoOutput, tc.AgentError)
			env.On("RunCommand", "git", []string{"config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`}).
				Return(tc.GitOutput, tc.GitError)

			origCache := template.Cache
			t.Cleanup(func() { template.Cache = origCache })
			template.Cache = &cache.Template{}
			template.Init(env, nil, nil)

			opts := options.Map{GPGAgentFetchKeys: tc.FetchKeys}
			maps.Copy(opts, tc.Options)

			agent := &GPGAgent{}
			agent.Init(opts, env)

			assert.True(t, agent.Enabled())

			if !tc.FetchKeys {
				env.AssertNotCalled(t, "RunCommand", "gpg-connect-agent", gpgKeyInfoArgs)
			}

			assert.Equal(t, tc.ExpectedRunning, agent.Running)
			assert.Equal(t, tc.ExpectedCount, agent.Count)
			assert.Equal(t, tc.ExpectedCachedCount, agent.CachedCount)
			assert.Equal(t, tc.ExpectedKeys, agent.Keys)
			assert.Equal(t, tc.ExpectedSigningKey, agent.SigningKey)
			assert.Equal(t, tc.ExpectedSignCommits, agent.SignCommits)
			assert.Equal(t, tc.ExpectedTTYSet, agent.TTYSet)
			assert.Equal(t, tc.ExpectedSSHSupport, agent.SSHSupport)
		})
	}
}

func TestGPGAgentCachesAgentAnswer(t *testing.T) {
	cache.Device.DeleteAll()
	cache.Session.DeleteAll()
	t.Cleanup(func() {
		cache.Device.DeleteAll()
		cache.Session.DeleteAll()
	})

	env := new(mock.Environment)
	env.On("Getenv", "GPG_TTY").Return("")
	env.On("Getenv", "SSH_AUTH_SOCK").Return("")
	env.On("Pwd").Return("/home/user/project")
	env.On("HasCommand", "gpg-connect-agent").Return(true)
	env.On("HasCommand", "git").Return(true)
	env.On("RunCommand", "gpg-connect-agent", gpgKeyInfoArgs).
		Return(gpgKeyInfo("S KEYINFO "+gpgGripA+" D - - 1 P - 600 -"), nil)
	env.On("RunCommand", "git", []string{"config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`}).
		Return("", gpgExitError(1))
	env.On("Shell").Return("bash")

	origCache := template.Cache
	t.Cleanup(func() { template.Cache = origCache })
	template.Cache = &cache.Template{}
	template.Init(env, nil, nil)

	for range 3 {
		agent := &GPGAgent{}
		agent.Init(options.Map{GPGAgentFetchKeys: true}, env)

		assert.True(t, agent.Enabled())
		assert.Equal(t, 1, agent.CachedCount)
	}

	env.AssertNumberOfCalls(t, "RunCommand", 2)
}

func TestGPGAgentCachesTheRunningState(t *testing.T) {
	cache.Device.DeleteAll()
	cache.Session.DeleteAll()
	t.Cleanup(func() {
		cache.Device.DeleteAll()
		cache.Session.DeleteAll()
	})

	env := new(mock.Environment)
	env.On("Getenv", "GPG_TTY").Return("")
	env.On("Getenv", "SSH_AUTH_SOCK").Return("")
	env.On("Pwd").Return("/home/user/project")
	env.On("HasCommand", "gpg-connect-agent").Return(true)
	env.On("HasCommand", "git").Return(true)
	env.On("RunCommand", "gpg-connect-agent", gpgInfoArgs).Return(gpgAgentInfo, nil)
	env.On("RunCommand", "git", []string{"config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`}).
		Return("", gpgExitError(1))
	env.On("Shell").Return("bash")

	origCache := template.Cache
	t.Cleanup(func() { template.Cache = origCache })
	template.Cache = &cache.Template{}
	template.Init(env, nil, nil)

	for range 3 {
		agent := &GPGAgent{}
		agent.Init(options.Map{}, env)

		assert.True(t, agent.Enabled())
		assert.True(t, agent.Running)
	}

	env.AssertNumberOfCalls(t, "RunCommand", 2)
}

func TestGPGAgentLogsUnexpectedGitFailures(t *testing.T) {
	cases := []struct {
		GitError       error
		Case           string
		ExpectedLogged bool
	}{
		{Case: "no signing configuration is expected", GitError: gpgExitError(1)},
		{Case: "an unsafe repository is not", GitError: gpgExitError(128), ExpectedLogged: true},
		{Case: "an invalid config file is not", GitError: gpgExitError(3), ExpectedLogged: true},
		{Case: "a failure without an exit code is not", GitError: errors.New("git exploded"), ExpectedLogged: true},
	}

	// log.Enable has no counterpart, so logging stays on for the rest of the package and the shared
	// buffer is read from the offset each case starts at.
	log.Enable(true)

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			cache.Device.DeleteAll()
			cache.Session.DeleteAll()
			t.Cleanup(func() {
				cache.Device.DeleteAll()
				cache.Session.DeleteAll()
			})

			env := new(mock.Environment)
			env.On("Getenv", "GPG_TTY").Return("")
			env.On("Getenv", "SSH_AUTH_SOCK").Return("")
			env.On("Pwd").Return("/home/user/project")
			env.On("Shell").Return("bash")
			env.On("HasCommand", "gpg-connect-agent").Return(true)
			env.On("HasCommand", "git").Return(true)
			env.On("RunCommand", "gpg-connect-agent", gpgInfoArgs).Return(gpgAgentInfo, nil)
			env.On("RunCommand", "git", []string{"config", "--get-regexp", `^(user\.signingkey|commit\.gpgsign)$`}).
				Return("", tc.GitError)

			origCache := template.Cache
			t.Cleanup(func() { template.Cache = origCache })
			template.Cache = &cache.Template{}
			template.Init(env, nil, nil)

			offset := len(log.String())

			agent := &GPGAgent{}
			agent.Init(options.Map{}, env)

			assert.True(t, agent.Enabled())
			assert.Empty(t, agent.SigningKey)
			assert.False(t, agent.SignCommits)
			assert.Equal(t, tc.ExpectedLogged, strings.Contains(log.String()[offset:], tc.GitError.Error()))
		})
	}
}
