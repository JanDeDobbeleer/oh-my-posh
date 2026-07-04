package cache

import (
	"hash/fnv"
	"os"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

// commandPathKeyPrefix namespaces persisted command lookups in the session
// store so they can't collide with unrelated cache keys.
const commandPathKeyPrefix = "command_path_"

// pathEnvHash returns an FNV-1a hash of the environment that determines how
// exec.LookPath resolves a command: PATH, plus PATHEXT on Windows (unset and
// therefore a no-op elsewhere). It's computed once per process -- the
// environment can't change under us in a short-lived prompt render.
var pathEnvHash = sync.OnceValue(func() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(os.Getenv("PATH")))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(os.Getenv("PATHEXT")))
	return h.Sum64()
})

// Command lookup TTLs for the session-persisted layer (L2). The in-memory
// layer (L1, Commands below) is unbounded for the lifetime of the process,
// matching existing behavior.
const (
	CommandPathTTL         = Duration("4h")
	CommandPathNegativeTTL = Duration("5m")
)

// commandPathEntry is the value persisted in the session store for a single
// command lookup. Found distinguishes a cached "does not exist" result
// (negative cache) from a resolved path. PathHash records the PATH(+PATHEXT)
// environment the lookup was resolved under; entries from a different
// environment are treated as misses so a PATH change mid-session (nvm use,
// venv activate, installing a tool earlier in PATH, ...) re-resolves
// immediately instead of serving a stale result for the TTL duration.
// Storing the hash in the entry (not the key) means a PATH change overwrites
// the entry in place rather than accumulating dead keys in the session store.
type commandPathEntry struct {
	Path     string
	PathHash uint64
	Found    bool
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
// picked up quickly. Entries are tagged with the current PATH(+PATHEXT)
// hash so they're only served back under the same lookup environment.
func PersistCommandPath(command, path string, found bool) {
	entry := commandPathEntry{Path: path, PathHash: pathEnvHash(), Found: found}

	ttl := CommandPathTTL
	if !found {
		ttl = CommandPathNegativeTTL
	}

	Set(Session, commandPathKey(command), entry, ttl)
}

// GetPersistedCommandPath returns a previously persisted command lookup
// result from the session store, if any. An entry persisted under a
// different PATH(+PATHEXT) environment is treated as a miss so the caller
// falls through to a fresh exec.LookPath (which overwrites the entry with
// the current environment's hash).
func GetPersistedCommandPath(command string) (path string, found, ok bool) {
	entry, exists := Get[commandPathEntry](Session, commandPathKey(command))
	if !exists {
		return "", false, false
	}

	if entry.PathHash != pathEnvHash() {
		return "", false, false
	}

	return entry.Path, entry.Found, true
}
