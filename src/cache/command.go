package cache

import (
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

// commandPathKeyPrefix namespaces persisted command lookups in the session
// store so they can't collide with unrelated cache keys.
const commandPathKeyPrefix = "command_path_"

// Command lookup TTLs for the session-persisted layer (L2). The in-memory
// layer (L1, Commands below) is unbounded for the lifetime of the process,
// matching existing behavior.
const (
	CommandPathTTL         = Duration("4h")
	CommandPathNegativeTTL = Duration("5m")
)

// commandPathEntry is the value persisted in the session store for a single
// command lookup. Found distinguishes a cached "does not exist" result
// (negative cache) from a resolved path.
type commandPathEntry struct {
	Path  string
	Found bool
}

type Command struct {
	// Commands is the in-memory (L1) cache: unchanged semantics, scoped to
	// the current process only.
	Commands *maps.Concurrent[string]
}

func commandPathKey(command string) string {
	return commandPathKeyPrefix + command
}

// Set stores a resolved command path in the in-memory (L1) cache only. This
// keeps existing in-process callers/semantics identical.
func (c *Command) Set(command, path string) {
	c.Commands.Set(command, path)
}

// Get looks up a command path in the in-memory (L1) cache only.
func (c *Command) Get(command string) (string, bool) {
	cacheCommand, found := c.Commands.Get(command)
	if !found {
		return "", false
	}

	return cacheCommand, true
}

// Persist stores a positive (path resolved) or negative (command not found)
// command lookup in the session (L2) store so subsequent prompt processes in
// the same session skip re-running exec.LookPath. Positive entries use a
// modest TTL and are revalidated with a cheap os.Stat on read; negative
// entries use a much shorter TTL so a command installed mid-session is
// picked up quickly.
func PersistCommandPath(command, path string, found bool) {
	entry := commandPathEntry{Path: path, Found: found}

	ttl := CommandPathTTL
	if !found {
		ttl = CommandPathNegativeTTL
	}

	Set(Session, commandPathKey(command), entry, ttl)
}

// GetPersistedCommandPath returns a previously persisted command lookup
// result from the session store, if any.
func GetPersistedCommandPath(command string) (path string, found, ok bool) {
	entry, exists := Get[commandPathEntry](Session, commandPathKey(command))
	if !exists {
		return "", false, false
	}

	return entry.Path, entry.Found, true
}
