package segments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodexStatusFromLocalSessions(t *testing.T) {
	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, ".codex")
	sessionRoot := filepath.Join(codexHome, "sessions", "2026", "06", "09")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(codexHome, "config.toml"),
		[]byte("model = \"gpt-5.5\"\nmodel_reasoning_effort = \"high\"\napproval_policy = \"never\"\nsandbox_mode = \"workspace-write\"\n"),
		0o644,
	))

	older := filepath.Join(sessionRoot, "rollout-2026-06-09T10-00-00-older-session.jsonl")
	writeCodexSessionFile(t, older, "older-session", "2026-06-09T14:00:00Z", 10, 20)
	newer := filepath.Join(sessionRoot, "rollout-2026-06-09T11-00-00-newer-session.jsonl")
	writeCodexSessionFile(t, newer, "newer-session", "2026-06-09T15:00:00Z", 30, 40)

	require.NoError(t, os.Chtimes(older, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)))
	require.NoError(t, os.Chtimes(newer, time.Now(), time.Now()))

	payload, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		CodexHome:   codexHome,
		SessionRoot: filepath.Join(codexHome, "sessions"),
	})
	require.NoError(t, err)

	assert.Equal(t, "token_count", payload.Type)
	assert.Equal(t, "newer-session", payload.ThreadID)
	assert.Equal(t, "C:/repo/newer-session", payload.CWD)
	assert.Equal(t, "0.138.0", payload.Version)
	assert.Equal(t, "high", payload.ReasoningEffort)
	assert.Equal(t, "never", payload.ApprovalMode)
	assert.Equal(t, "workspace-write", payload.SandboxPolicy)
	assert.Equal(t, "gpt-5.5", payload.Model.DisplayName)
	assert.Equal(t, 40, payload.Info.TotalTokenUsage.TotalTokens)
}

func TestCodexStatusFromLocalSessionsUsesRequestedSession(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))

	older := filepath.Join(sessionRoot, "rollout-2026-06-09T10-00-00-older-session.jsonl")
	writeCodexSessionFile(t, older, "older-session", "2026-06-09T14:00:00Z", 10, 20)
	newer := filepath.Join(sessionRoot, "rollout-2026-06-09T11-00-00-newer-session.jsonl")
	writeCodexSessionFile(t, newer, "newer-session", "2026-06-09T15:00:00Z", 30, 40)

	require.NoError(t, os.Chtimes(older, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)))
	require.NoError(t, os.Chtimes(newer, time.Now(), time.Now()))

	payload, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
		SessionID:   "older-session",
	})
	require.NoError(t, err)

	assert.Equal(t, "older-session", payload.ThreadID)
	assert.Equal(t, 20, payload.Info.TotalTokenUsage.TotalTokens)
}

func TestCodexStatusFromLocalSessionsSkipsMismatchedRequestedSession(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))

	mismatch := filepath.Join(sessionRoot, "rollout-2026-06-09T10-00-00-requested-session.jsonl")
	writeCodexSessionFile(t, mismatch, "different-session", "2026-06-09T14:00:00Z", 10, 20)

	_, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
		SessionID:   "requested-session",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, errCodexNoTokenCount)
}

func TestCodexStatusFromLocalSessionsReturnsErrorWhenNoTokenCount(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionRoot, "rollout-no-token.jsonl"),
		[]byte("{\"timestamp\":\"2026-06-09T14:00:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"agent_message\",\"message\":\"done\"}}\n"),
		0o644,
	))

	_, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, errCodexNoTokenCount)
}

func TestCodexStatusFromLocalSessionsReturnsMalformedJSONLError(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionRoot, "rollout-malformed.jsonl"),
		[]byte("{\"timestamp\":\"2026-06-09T14:00:00Z\",\"type\":\"event_msg\",\"payload\":"),
		0o644,
	))

	_, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, errCodexNoTokenCount)
}

func TestCodexStatusFromLocalSessionsIgnoresNonDateDirectories(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(filepath.Join(sessionRoot, "archive"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sessionRoot, "2026", "06", "09"), 0o755))

	session := filepath.Join(sessionRoot, "2026", "06", "09", "rollout-2026-06-09T10-00-00-session.jsonl")
	writeCodexSessionFile(t, session, "session", "2026-06-09T14:00:00Z", 10, 20)

	payload, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
	})
	require.NoError(t, err)
	assert.Equal(t, "session", payload.ThreadID)
}

func TestCodexSegmentDiscoversLocalStatus(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, ".codex")
	sessionRoot := filepath.Join(codexHome, "sessions", "2026", "06", "09")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))
	writeCodexSessionFile(t, filepath.Join(sessionRoot, "rollout-local-session.jsonl"), "local-session", "2026-06-09T14:00:00Z", 10, 20)

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("")
	env.On("Getenv", defaultCodexStatusJSONEnv).Return("")
	env.On("Getenv", "CODEX_HOME").Return(codexHome)
	env.On("Home").Return(tmp)

	segment := &Codex{
		Base: Base{
			env:     env,
			options: options.Map{},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "local-session", segment.ThreadID)
	assert.Equal(t, 20, segment.Info.TotalTokenUsage.TotalTokens)

	cache.Delete(cache.Session, cache.CODEXCACHE)
}

func TestCodexCacheKeyUsesStatusSource(t *testing.T) {
	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("C:/tmp/codex-status.json")

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				codexHome:      "C:/Users/Test/.codex",
				codexSessionID: "thread-1",
			},
		},
	}

	key, ok := segment.CacheKey()
	require.True(t, ok)

	expectedKey := "codex|discover|true|file-env|POSH_CODEX_STATUS_FILE|" +
		"file|C:/tmp/codex-status.json|json-env|POSH_CODEX_STATUS|" +
		"session|thread-1|root|" + filepath.Join("C:/Users/Test/.codex", "sessions")
	assert.Equal(
		t,
		expectedKey,
		key,
	)
}

func writeCodexSessionFile(t *testing.T, path, sessionID, timestamp string, firstTokens, secondTokens int) {
	t.Helper()

	content := "" +
		"{\"timestamp\":\"" + timestamp + "\",\"type\":\"session_meta\",\"payload\":{\"id\":\"" + sessionID + "\",\"cwd\":\"C:/repo/" + sessionID + "\",\"cli_version\":\"0.138.0\"}}\n" +
		codexTokenCountLine(timestamp, firstTokens) + "\n" +
		codexTokenCountLine(timestamp, secondTokens) + "\n"

	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func codexTokenCountLine(timestamp string, totalTokens int) string {
	return fmt.Sprintf(
		`{"timestamp":"%s","type":"event_msg","payload":{"type":"token_count",`+
			`"info":{"total_token_usage":{"total_tokens":%d},"last_token_usage":{"total_tokens":100},`+
			`"model_context_window":10000},"rate_limits":{"primary":{"used_percent":12},`+
			`"secondary":{"used_percent":34}}}}`,
		timestamp,
		totalTokens,
	)
}
