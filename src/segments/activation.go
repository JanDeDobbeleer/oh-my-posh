package segments

// Activation describes the cheap preconditions under which a segment can
// possibly be enabled. Conditions are OR'd: any match keeps the segment
// alive. The zero value (or Always) means "no gate - always execute".
type Activation struct {
	// FileGlobs: the segment can only activate when the current working
	// directory contains a file matching one of these globs (evaluated via
	// env.HasFiles, which reads from the cached directory listing).
	FileGlobs []string
	Always    bool
}

// Activator is an optional interface a segment writer can implement to let
// the engine skip the (potentially expensive) Enabled() probe when the
// segment provably cannot activate in the current working directory.
// Activation() is called after Init(), so implementations may consult their
// options. Writers that do not implement it always execute.
type Activator interface {
	Activation() Activation
}
