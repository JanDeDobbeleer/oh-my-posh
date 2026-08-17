package runtime

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerminalDataOnlyRefusesEveryProbe is the guarantee the studio rests on:
// with DataOnly set, nothing this environment exposes can reach the machine.
// It is asserted here, at the environment, rather than per segment, because a
// segment only has to call one of these to escape - git's StashCount reads
// logs/refs/stash from a method the template invokes, long after the recorded
// data has been restored, and no rule about recorded data would have caught it.
//
// Every case names a real path or command that exists on the machine running
// the test, so a passing assertion means the call was refused rather than
// merely finding nothing.
func TestTerminalDataOnlyRefusesEveryProbe(t *testing.T) {
	term := &Terminal{CmdFlags: &Flags{DataOnly: true, PWD: "/somewhere"}}
	term.Init(term.CmdFlags)

	t.Run("FileContent", func(t *testing.T) {
		assert.Empty(t, term.FileContent("go.mod"))
	})

	t.Run("HasFiles", func(t *testing.T) {
		assert.False(t, term.HasFiles("*.go"))
	})

	t.Run("HasFilesInDir", func(t *testing.T) {
		assert.False(t, term.HasFilesInDir(".", "*.go"))
	})

	t.Run("HasFolder", func(t *testing.T) {
		assert.False(t, term.HasFolder("."))
	})

	t.Run("LsDir", func(t *testing.T) {
		assert.Empty(t, term.LsDir("."))
	})

	t.Run("HasCommand", func(t *testing.T) {
		assert.False(t, term.HasCommand("go"))
	})

	t.Run("CommandPath", func(t *testing.T) {
		assert.Empty(t, term.CommandPath("go"))
	})

	t.Run("StatFile", func(t *testing.T) {
		stat, err := term.StatFile(os.Args[0])
		assert.Equal(t, FileStat{}, stat)
		require.ErrorIs(t, err, errDataOnly)
	})

	t.Run("RunCommand", func(t *testing.T) {
		out, err := term.RunCommand("go", "version")
		assert.Empty(t, out)
		require.ErrorIs(t, err, errDataOnly)
	})

	t.Run("HasParentFilePath", func(t *testing.T) {
		info, err := term.HasParentFilePath("go.mod", false)
		assert.Nil(t, info)
		require.ErrorIs(t, err, errDataOnly)
	})

	t.Run("HTTPRequest", func(t *testing.T) {
		body, err := term.HTTPRequest("https://ohmyposh.dev", nil, 1000)
		assert.Nil(t, body)
		require.ErrorIs(t, err, errDataOnly)
	})

	// Not a machine probe in the same sense, but the same divergence risk: a
	// browser has no environment, so reading the real one would make the CLI
	// and the wasm build disagree from identical inputs.
	t.Run("Getenv", func(t *testing.T) {
		t.Setenv("OMP_DATAONLY_PROBE", "leaked")
		assert.Empty(t, term.Getenv("OMP_DATAONLY_PROBE"))
	})
}

// The same calls must work normally when the flag is off, or the guard above
// would be indistinguishable from the environment simply being broken.
func TestTerminalWithoutDataOnlyStillReaches(t *testing.T) {
	term := &Terminal{CmdFlags: &Flags{}}
	term.Init(term.CmdFlags)

	t.Setenv("OMP_DATAONLY_PROBE", "present")

	assert.Equal(t, "present", term.Getenv("OMP_DATAONLY_PROBE"))
	assert.True(t, term.HasCommand("go"), "the toolchain running this test is on PATH")
}
