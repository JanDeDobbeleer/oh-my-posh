package segments

import "github.com/jandedobbeleer/oh-my-posh/src/runtime"

// Activation aliases runtime.Activation, where the type (and its evaluation
// contract) lives so the config package can reference it in the
// SegmentWriter interface without importing this package. Writers implement
// Activation() to declare their preconditions; Base provides the ungated
// default, so a writer only writes code when it has a real gate.
type Activation = runtime.Activation
