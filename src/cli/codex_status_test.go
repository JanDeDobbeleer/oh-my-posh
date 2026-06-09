package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	data, err := codexStatusFromLocalSessions(codexLocalStatusOptions{
		CodexHome:   codexHome,
		SessionRoot: filepath.Join(codexHome, "sessions"),
	})
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(data, &payload))
	assert.Equal(t, "token_count", payload["type"])
	assert.Equal(t, "newer-session", payload["thread_id"])
	assert.Equal(t, "C:/repo/newer-session", payload["cwd"])
	assert.Equal(t, "0.138.0", payload["version"])
	assert.Equal(t, "high", payload["reasoning_effort"])
	assert.Equal(t, "never", payload["approval_mode"])
	assert.Equal(t, "workspace-write", payload["sandbox_policy"])
	assert.Equal(t, float64(40), payload["sequence"])
	assert.Equal(t, map[string]any{
		"id":           "gpt-5.5",
		"display_name": "gpt-5.5",
	}, payload["model"])
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

	data, err := codexStatusFromLocalSessions(codexLocalStatusOptions{
		SessionRoot: sessionRoot,
		SessionID:   "older-session",
	})
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(data, &payload))
	assert.Equal(t, "older-session", payload["thread_id"])
	assert.Equal(t, float64(20), payload["sequence"])
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

	_, err := codexStatusFromLocalSessions(codexLocalStatusOptions{
		SessionRoot: sessionRoot,
	})
	require.Error(t, err)
}

func writeCodexSessionFile(t *testing.T, path, sessionID, timestamp string, firstSequence, secondSequence int) {
	t.Helper()

	content := "" +
		"{\"timestamp\":\"" + timestamp + "\",\"type\":\"session_meta\",\"payload\":{\"id\":\"" + sessionID + "\",\"cwd\":\"C:/repo/" + sessionID + "\",\"cli_version\":\"0.138.0\"}}\n" +
		codexTokenCountLine(timestamp, firstSequence) + "\n" +
		codexTokenCountLine(timestamp, secondSequence) + "\n"

	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func codexTokenCountLine(timestamp string, sequence int) string {
	return fmt.Sprintf(
		`{"timestamp":"%s","type":"event_msg","payload":{"type":"token_count","sequence":%d,`+
			`"info":{"total_token_usage":{"total_tokens":1000},"last_token_usage":{"total_tokens":100},`+
			`"model_context_window":10000},"rate_limits":{"primary":{"used_percent":12},`+
			`"secondary":{"used_percent":34}}}}`,
		timestamp,
		sequence,
	)
}
