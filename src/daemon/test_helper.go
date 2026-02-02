package daemon

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	windowsOS = "windows"
)

func setTestEnv(t *testing.T, tmpDir string) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", tmpDir)
	if runtime.GOOS == windowsOS {
		t.Setenv("LOCALAPPDATA", tmpDir)
	}
}

func createTestConfig(t *testing.T) string {
	t.Helper()
	content := `{
		"version": 4,
		"blocks": [
			{
				"type": "prompt",
				"alignment": "left",
				"segments": [
					{
						"type": "text",
						"style": "plain",
						"template": "hello"
					}
				]
			}
		]
	}`
	tmpFile, err := os.CreateTemp("", "omp-config-*.json")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}
