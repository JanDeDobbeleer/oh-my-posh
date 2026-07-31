package config

// DSCTracker records which config source files were parsed, for the CLI's
// "config dsc" resource. It is CLI-only bookkeeping: cli/dsc registers the
// real implementation via NewDSCTracker; the wasm render build never imports
// cli, so newDSCTracker stays nil there and Parse's bookkeeping is skipped -
// keeping invopop/jsonschema (go/ast, go/parser, go/doc) and the command tree/
// flag machinery, which the real dsc.Resource pulls in, out of the wasm link
// graph. ParseBytes/ParseData
// (the entry points wasm actually uses) never call this at all.
type DSCTracker interface {
	Load()
	Add(configFile string)
	Save()
}

// NewDSCTracker is set by cli/dsc's init().
var NewDSCTracker func() DSCTracker

func newDSCTracker() DSCTracker {
	if NewDSCTracker == nil {
		return nil
	}

	return NewDSCTracker()
}
