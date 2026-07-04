package cache

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/stretchr/testify/assert"
)

func TestPersistCommandPathRoundTrip(t *testing.T) {
	session = Session.new()

	PersistCommandPath("git", "/usr/bin/git", true)

	path, found, ok := GetPersistedCommandPath("git")
	assert.True(t, ok)
	assert.True(t, found)
	assert.Equal(t, "/usr/bin/git", path)
}

func TestPersistCommandPathNegative(t *testing.T) {
	session = Session.new()

	PersistCommandPath("does-not-exist", "", false)

	path, found, ok := GetPersistedCommandPath("does-not-exist")
	assert.True(t, ok)
	assert.False(t, found)
	assert.Empty(t, path)
}

func TestGetPersistedCommandPathMiss(t *testing.T) {
	session = Session.new()

	_, _, ok := GetPersistedCommandPath("never-set")
	assert.False(t, ok)
}

func TestGetPersistedCommandPathStalePathHashIsAMiss(t *testing.T) {
	session = Session.new()

	// Simulate an entry persisted under a different PATH environment (or a
	// pre-PathHash entry, which decodes with PathHash == 0).
	entry := commandPathEntry{Path: "/usr/bin/git", PathHash: pathEnvHash() + 1, Found: true}
	Set(Session, commandPathKey("git"), entry, CommandPathTTL)

	_, _, ok := GetPersistedCommandPath("git")
	assert.False(t, ok, "entry persisted under a different PATH must be treated as a miss")
}

func TestCommandPathKeyIsPrefixed(t *testing.T) {
	assert.Equal(t, "command_path_git", commandPathKey("git"))
}

func TestCommandL1CacheUnaffectedBySessionStore(t *testing.T) {
	session = Session.new()

	c := &Command{Commands: maps.NewConcurrent[string]()}
	c.Set("git", "/usr/bin/git")

	path, ok := c.Get("git")
	assert.True(t, ok)
	assert.Equal(t, "/usr/bin/git", path)

	// L1 lookups must not consult/require the session store.
	_, found, ok := GetPersistedCommandPath("git")
	assert.False(t, ok)
	assert.False(t, found)
}
