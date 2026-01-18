package daemon

import (
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// Environment wraps runtime.Terminal to use environment variables
// from the client request instead of the daemon's own environment.
// This ensures segments see the shell's environment, not the daemon's.
type Environment struct {
	*runtime.Terminal
	envVars map[string]string
}

// NewEnvironment creates a new daemon environment.
// The envVars map contains environment variables from the client request.
func NewEnvironment(flags *runtime.Flags, envVars map[string]string) *Environment {
	term := &runtime.Terminal{}
	term.Init(flags)

	return &Environment{
		Terminal: term,
		envVars:  envVars,
	}
}

// Getenv returns the value of the environment variable named by the key.
// It first checks the request-specific env vars, then falls back to the
// daemon's own environment.
func (de *Environment) Getenv(key string) string {
	// Check request-specific env vars first
	if val, ok := de.envVars[key]; ok {
		log.Debugf("daemon env (from request): %s=%s", key, val)
		return val
	}

	// Fall back to daemon's environment
	return de.Terminal.Getenv(key)
}
