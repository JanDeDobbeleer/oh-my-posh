package segments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

	older := filepath.Join(sessionRoot, "rollout-2026-06-09T10-00-00.jsonl")
	writeCodexSessionFile(t, older, "older-session", "2026-06-09T14:00:00Z", 10, 20)
	newer := filepath.Join(sessionRoot, "rollout-2026-06-09T11-00-00-newer-session.jsonl")
	writeCodexSessionFile(t, newer, "newer-session", "2026-06-09T15:00:00Z", 30, 40)

	payload, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
		SessionID:   "older-session",
	})
	require.NoError(t, err)

	assert.Equal(t, "older-session", payload.ThreadID)
	assert.Equal(t, 20, payload.Info.TotalTokenUsage.TotalTokens)
}

func TestCodexStatusFromLocalSessionsConsidersRootFilesAfterDateCap(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	dateRoot := filepath.Join(sessionRoot, "2026", "06", "09")
	require.NoError(t, os.MkdirAll(dateRoot, 0o755))

	for i := 0; i < codexMaxSessionFiles; i++ {
		sessionID := fmt.Sprintf("dated-session-%02d", i)
		path := filepath.Join(dateRoot, fmt.Sprintf("rollout-2026-06-09T10-%02d-00-%s.jsonl", i, sessionID))
		writeCodexSessionFile(t, path, sessionID, "2026-06-09T14:00:00Z", 10, 20)
	}

	writeCodexSessionFile(
		t,
		filepath.Join(sessionRoot, "rollout-2026-06-10T10-00-00-root-session.jsonl"),
		"root-session",
		"2026-06-10T14:00:00Z",
		30,
		40,
	)

	payload, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: sessionRoot,
	})
	require.NoError(t, err)
	assert.Equal(t, "root-session", payload.ThreadID)
	assert.Equal(t, 40, payload.Info.TotalTokenUsage.TotalTokens)
}

func TestCodexSessionFilesCapsRequestedSessionCandidates(t *testing.T) {
	tmp := t.TempDir()
	sessionRoot := filepath.Join(tmp, "sessions")
	require.NoError(t, os.MkdirAll(sessionRoot, 0o755))

	for i := 0; i < codexMaxRequestedSessionFiles+10; i++ {
		path := filepath.Join(sessionRoot, fmt.Sprintf("rollout-2026-06-09T10-%03d-00.jsonl", i))
		writeCodexSessionFile(t, path, fmt.Sprintf("session-%03d", i), "2026-06-09T14:00:00Z", 10, 20)
	}

	files, err := codexSessionFiles(sessionRoot, "requested-session")
	require.NoError(t, err)
	require.Len(t, files, codexMaxRequestedSessionFiles)
}

func TestSortCodexSessionFilesBreaksTiesByPath(t *testing.T) {
	files := []codexSessionFile{
		{path: filepath.Join("sessions", "2026", "06", "09", "rollout.jsonl"), sortKey: "rollout.jsonl"},
		{path: filepath.Join("sessions", "rollout.jsonl"), sortKey: "rollout.jsonl"},
	}

	sortCodexSessionFiles(files)

	assert.Equal(t, filepath.Join("sessions", "rollout.jsonl"), files[0].path)
}

func TestCodexStatusFromLocalSessionsReturnsActionableErrorWhenSessionRootMissing(t *testing.T) {
	_, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "set CODEX_HOME")
	assert.Contains(t, err.Error(), "--session-root")
}

func TestCodexStatusFromLocalSessionsReturnsActionableErrorWhenNoSessionFiles(t *testing.T) {
	tmp := t.TempDir()

	_, err := CodexStatusFromLocalSessions(CodexLocalStatusOptions{
		SessionRoot: tmp,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start Codex once")
	assert.Contains(t, err.Error(), "--codex-home")
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

	statusFileHash := codexCacheKeyHash("C:/tmp/codex-status.json")
	sessionRootHash := codexCacheKeyHash(filepath.Join("C:/Users/Test/.codex", "sessions"))
	expectedKey := "codex|discover|true|file-env|POSH_CODEX_STATUS_FILE|" +
		"file|" + statusFileHash + "|json-env|POSH_CODEX_STATUS|" +
		"session|thread-1|root|" + sessionRootHash
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
