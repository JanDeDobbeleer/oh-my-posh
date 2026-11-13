package runtime

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalHostName(t *testing.T) {
	hostName := "hello"
	assert.Equal(t, hostName, cleanHostName(hostName))
}

func TestHostNameWithLocal(t *testing.T) {
	hostName := "hello.local"
	assert.Equal(t, "hello", cleanHostName(hostName))
}

func TestHostNameWithLan(t *testing.T) {
	hostName := "hello.lan"
	cleanHostName := cleanHostName(hostName)
	assert.Equal(t, "hello", cleanHostName)
}

func TestDirMatchesOneOf(t *testing.T) {
	cases := []struct {
		GOOS     string
		HomeDir  string
		Dir      string
		Pattern  string
		Expected bool
	}{
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill", Pattern: "/home/bill", Expected: true},
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/foo", Expected: true},
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/Foo", Expected: false},
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~\\\\foo", Expected: true},
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill/foo/bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: LINUX, HomeDir: "/home/bill", Dir: "/home/bill/foo", Pattern: "~/fo\\w", Expected: true},

		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:\\\\Users\\\\Bill", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "C:/Users/Bill", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "c:/users/bill", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill", Pattern: "~", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/Foo", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/foo", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo\\Bar", Pattern: "~/fo.*", Expected: true},
		{GOOS: WINDOWS, HomeDir: "C:\\Users\\Bill", Dir: "C:\\Users\\Bill\\Foo", Pattern: "~/fo\\w", Expected: true},
	}

	for _, tc := range cases {
		got := dirMatchesOneOf(tc.Dir, tc.HomeDir, tc.GOOS, []string{tc.Pattern})
		assert.Equal(t, tc.Expected, got)
	}
}

func TestDirMatchesOneOfRegexInverted(t *testing.T) {
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects[\\/]).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	_ = dirMatchesOneOf("Projects/oh-my-posh", "", LINUX, []string{"(?!Projects[\\/]).*"})
}

func TestDirMatchesOneOfRegexInvertedNonEscaped(t *testing.T) {
	// detect panic(thrown by MustCompile)
	defer func() {
		if err := recover(); err != nil {
			// display a message explaining omp failed(with the err)
			assert.Equal(t, "regexp: Compile(`^(?!Projects/).*$`): error parsing regexp: invalid or unsupported Perl syntax: `(?!`", err)
		}
	}()
	_ = dirMatchesOneOf("Projects/oh-my-posh", "", LINUX, []string{"(?!Projects/).*"})
}

func TestPromptCountSurvivesIdleSession(t *testing.T) {
	sessionFile := filepath.Join(cache.Path(), fmt.Sprintf("zsh.%s.%s", cache.SessionID(), cache.DeviceStore))

	renderPrompt := func(promptType string) int {
		cache.Init("zsh", cache.Persist)
		defer cache.Close()

		term := &Terminal{}
		term.Init(&Flags{Type: promptType})
		return term.CmdFlags.PromptCount
	}

	cases := []struct {
		Case string
		Idle time.Duration
	}{
		{Case: "idle for an hour", Idle: time.Hour},
		{Case: "idle for more than a day", Idle: 25 * time.Hour},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			_ = os.Remove(sessionFile)

			assert.Equal(t, 1, renderPrompt(PRIMARY))

			data, err := os.ReadFile(sessionFile)
			require.NoError(t, err)

			var entries maps.Simple[*cache.Entry[any]]
			require.NoError(t, gob.NewDecoder(bytes.NewReader(data)).Decode(&entries))

			for _, entry := range entries {
				entry.Timestamp -= int64(tc.Idle.Seconds())
			}

			var aged bytes.Buffer
			require.NoError(t, gob.NewEncoder(&aged).Encode(entries))
			require.NoError(t, os.WriteFile(sessionFile, aged.Bytes(), 0o600))

			assert.Equal(t, 1, renderPrompt("transient"), "the transient prompt keeps the count of the prompt it replaces")
			assert.Equal(t, 2, renderPrompt(PRIMARY), "the next primary prompt continues the count")
		})
	}
}
