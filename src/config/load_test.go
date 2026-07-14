package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeExtendsFixture(t *testing.T, path, contents string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
}

// TestParseExtendsChainThreeLevels proves that a multi-hop extends chain
// (A extends B extends C), with every config in the same directory, still
// merges every level correctly.
//
// The distinguishing fields are all strings: merge() only skips a field when
// it is the Go zero value (see isZeroValue in merge.go), and for bool/int
// fields the zero value (false/0) is indistinguishable from "not set in this
// hop's JSON", so a bool or int field would always be clobbered by a later,
// unset hop. String fields don't have that problem, so they reliably prove
// which hop contributed which value.
func TestParseExtendsChainThreeLevels(t *testing.T) {
	dir := t.TempDir()

	writeExtendsFixture(t, filepath.Join(dir, "c.omp.json"), `{
	"accent_color": "#ff0000"
}`)

	writeExtendsFixture(t, filepath.Join(dir, "b.omp.json"), `{
	"extends": "c.omp.json",
	"pwd": "/from/b"
}`)

	aPath := filepath.Join(dir, "a.omp.json")
	writeExtendsFixture(t, aPath, `{
	"extends": "b.omp.json",
	"console_title_template": "hello"
}`)

	cfg, err := Parse(aPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "#ff0000", string(cfg.AccentColor), "field from the deepest (C) config should be merged")
	assert.Equal(t, "/from/b", cfg.PWD, "field from the middle (B) config should be merged")
	assert.Equal(t, "hello", cfg.ConsoleTitleTemplate, "field from the top-level (A) config should be merged")
}

// TestParseExtendsRelativePathPerHop proves that a relative extends path
// declared inside a non-top-level config is resolved against that config's
// own directory, not the directory of the config that started the chain.
func TestParseExtendsRelativePathPerHop(t *testing.T) {
	dir := t.TempDir()

	// c.omp.json lives two directories deeper than a.omp.json, so resolving
	// b.omp.json's "extends" against a.omp.json's directory (the bug) would
	// look for it in the wrong place and fail to find it.
	writeExtendsFixture(t, filepath.Join(dir, "level1", "level2", "c.omp.json"), `{
	"accent_color": "#ff0000"
}`)

	writeExtendsFixture(t, filepath.Join(dir, "level1", "b.omp.json"), `{
	"extends": "level2/c.omp.json",
	"pwd": "/from/b"
}`)

	aPath := filepath.Join(dir, "a.omp.json")
	writeExtendsFixture(t, aPath, `{
	"extends": "level1/b.omp.json",
	"console_title_template": "hello"
}`)

	cfg, err := Parse(aPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "#ff0000", string(cfg.AccentColor), "C's relative extends should resolve against B's own directory")
	assert.Equal(t, "/from/b", cfg.PWD)
	assert.Equal(t, "hello", cfg.ConsoleTitleTemplate)
}

// TestParseExtendsCycleDoesNotHang proves that a circular extends chain
// (A extends B extends A) terminates instead of looping forever, and that
// the fields merged before the cycle was detected are still applied.
func TestParseExtendsCycleDoesNotHang(t *testing.T) {
	dir := t.TempDir()

	aPath := filepath.Join(dir, "a.omp.json")
	writeExtendsFixture(t, aPath, `{
	"extends": "b.omp.json",
	"console_title_template": "hello"
}`)

	writeExtendsFixture(t, filepath.Join(dir, "b.omp.json"), `{
	"extends": "a.omp.json",
	"pwd": "/from/b"
}`)

	type result struct {
		cfg *Config
		err error
	}

	done := make(chan result, 1)

	go func() {
		cfg, err := Parse(aPath)
		done <- result{cfg: cfg, err: err}
	}()

	select {
	case res := <-done:
		require.NoError(t, res.err)
		require.NotNil(t, res.cfg)

		assert.Equal(t, "hello", res.cfg.ConsoleTitleTemplate, "fields merged before the cycle was detected should survive")
		assert.Equal(t, "/from/b", res.cfg.PWD)
	case <-time.After(5 * time.Second):
		t.Fatal("Parse did not return, circular extends was not detected")
	}
}
