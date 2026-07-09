package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeDataFile writes a JSON template data file to a temp dir and returns
// its path.
func writeDataFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	return path
}

// withDataPath sets the package-level dataPath var for the duration of the
// test and restores it afterwards - dataPath is shared with the print/image
// commands so tests must not leak it across each other.
func withDataPath(t *testing.T, path string) {
	t.Helper()

	previous := dataPath
	dataPath = path

	t.Cleanup(func() { dataPath = previous })
}

// noneChanged reports that no CLI flag was explicitly set.
func noneChanged(string) bool { return false }

// changedSet returns a "Changed" func that reports true only for the given
// flag names, mimicking cmd.Flags().Changed for an explicitly-set flag.
func changedSet(names ...string) func(string) bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	return func(name string) bool { return set[name] }
}

func TestApplyDataFile_EmptyPathIsNoop(t *testing.T) {
	withDataPath(t, "")

	flags := &runtime.Flags{PWD: "/live/pwd", ErrorCode: 1, ExecutionTime: 2.5, PipeStatus: "0"}

	err := applyDataFile(flags, noneChanged)
	require.NoError(t, err)

	assert.Equal(t, "/live/pwd", flags.PWD)
	assert.Equal(t, 1, flags.ErrorCode)
	assert.InDelta(t, 2.5, flags.ExecutionTime, 0)
	assert.Equal(t, "0", flags.PipeStatus)
	assert.Nil(t, flags.EnvData)
	assert.Nil(t, flags.SegmentData)
}

func TestApplyDataFile_MissingFileErrors(t *testing.T) {
	withDataPath(t, filepath.Join(t.TempDir(), "does-not-exist.json"))

	flags := &runtime.Flags{}

	err := applyDataFile(flags, noneChanged)
	assert.Error(t, err)
}

func TestApplyDataFile_UnsupportedExtensionErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))
	withDataPath(t, path)

	flags := &runtime.Flags{}

	err := applyDataFile(flags, noneChanged)
	assert.Error(t, err)
}

func TestApplyDataFile_RoutesAllFourKeysWhenNoFlagsChanged(t *testing.T) {
	path := writeDataFile(t, `{
		"env": {"PWD": "/data/pwd", "Code": 3, "ExecutionTime": 12.5, "PipeStatus": "0 1"},
		"segments": {"az": {"Name": "my-sub"}}
	}`)
	withDataPath(t, path)

	flags := &runtime.Flags{PWD: "/live/pwd", ErrorCode: 1, ExecutionTime: 2.5, PipeStatus: "0"}

	err := applyDataFile(flags, noneChanged)
	require.NoError(t, err)

	assert.Equal(t, "/data/pwd", flags.PWD)
	assert.Equal(t, 3, flags.ErrorCode)
	assert.InDelta(t, 12.5, flags.ExecutionTime, 0)
	assert.Equal(t, "0 1", flags.PipeStatus)

	require.NotNil(t, flags.SegmentData)
	assert.JSONEq(t, `{"Name": "my-sub"}`, string(flags.SegmentData["az"]))

	require.NotNil(t, flags.EnvData)

	var env map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(flags.EnvData, &env))
	assert.Contains(t, env, "PWD")
}

func TestApplyDataFile_ExplicitCLIFlagWinsOverDataFile(t *testing.T) {
	path := writeDataFile(t, `{
		"env": {"PWD": "/data/pwd", "Code": 3, "ExecutionTime": 12.5, "PipeStatus": "0 1"}
	}`)
	withDataPath(t, path)

	flags := &runtime.Flags{PWD: "/live/pwd", ErrorCode: 1, ExecutionTime: 2.5, PipeStatus: "0"}

	err := applyDataFile(flags, changedSet("pwd", "status", "execution-time", "pipestatus"))
	require.NoError(t, err)

	// Every routed key was explicitly set on the CLI, so the live/flag
	// value must survive untouched despite the data file providing all four.
	assert.Equal(t, "/live/pwd", flags.PWD)
	assert.Equal(t, 1, flags.ErrorCode)
	assert.InDelta(t, 2.5, flags.ExecutionTime, 0)
	assert.Equal(t, "0", flags.PipeStatus)
}

func TestApplyDataFile_PerKeyPrecedence(t *testing.T) {
	path := writeDataFile(t, `{
		"env": {"PWD": "/data/pwd", "Code": 3, "ExecutionTime": 12.5, "PipeStatus": "0 1"}
	}`)
	withDataPath(t, path)

	flags := &runtime.Flags{PWD: "/live/pwd", ErrorCode: 1, ExecutionTime: 2.5, PipeStatus: "0"}

	// Only "status" is explicitly set on the CLI; the other three keys
	// should still be routed from the data file.
	err := applyDataFile(flags, changedSet("status"))
	require.NoError(t, err)

	assert.Equal(t, "/data/pwd", flags.PWD, "pwd not explicitly set, so the data file value should apply")
	assert.Equal(t, 1, flags.ErrorCode, "status explicitly set, so it must keep the live/flag value")
	assert.InDelta(t, 12.5, flags.ExecutionTime, 0, "execution-time not explicitly set, so the data file value should apply")
	assert.Equal(t, "0 1", flags.PipeStatus, "pipestatus not explicitly set, so the data file value should apply")
}

func TestApplyDataFile_MissingEnvKeysLeaveFlagsUntouched(t *testing.T) {
	path := writeDataFile(t, `{"segments": {"az": {"Name": "my-sub"}}}`)
	withDataPath(t, path)

	flags := &runtime.Flags{PWD: "/live/pwd", ErrorCode: 1, ExecutionTime: 2.5, PipeStatus: "0"}

	err := applyDataFile(flags, noneChanged)
	require.NoError(t, err)

	assert.Equal(t, "/live/pwd", flags.PWD)
	assert.Equal(t, 1, flags.ErrorCode)
	assert.InDelta(t, 2.5, flags.ExecutionTime, 0)
	assert.Equal(t, "0", flags.PipeStatus)
}
