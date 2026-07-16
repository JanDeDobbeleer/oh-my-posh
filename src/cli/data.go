package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// dataPath is the shared --data flag value used by printCmd and imageCmd to
// render deterministically from a recorded template data file instead of
// the live environment.
var dataPath string

// applyDataFile loads the template data file referenced by --data (if any)
// and routes it onto flags. It is a no-op when dataPath is empty.
//
// SegmentData and EnvData are copied onto flags verbatim - the rest of the
// pipeline (segment replay, template cache overlay) consumes them from
// there. The env keys that map directly onto runtime.Flags fields (see
// routedEnvDataKeys in the template package) are routed here explicitly,
// with precedence explicit CLI flag > data file > live environment: changed
// reports whether the corresponding CLI flag (by name) was set explicitly,
// in which case the data file's value is skipped.
func applyDataFile(flags *runtime.Flags, changed func(name string) bool) error {
	if dataPath == "" {
		return nil
	}

	data, err := config.LoadData(dataPath)
	if err != nil {
		return err
	}

	flags.SegmentData = data.Segments
	flags.EnvData = data.Env

	envFlags, err := data.EnvFlags()
	if err != nil {
		return err
	}

	if envFlags.PWD != nil && !changed("pwd") {
		flags.PWD = *envFlags.PWD
	}

	if envFlags.Code != nil && !changed("status") {
		flags.ErrorCode = *envFlags.Code
	}

	if envFlags.ExecutionTime != nil && !changed("execution-time") {
		flags.ExecutionTime = *envFlags.ExecutionTime
	}

	if envFlags.PipeStatus != nil && !changed("pipestatus") {
		flags.PipeStatus = *envFlags.PipeStatus
	}

	if envFlags.Interrupted != nil && !changed("interrupted") {
		flags.Interrupted = *envFlags.Interrupted
	}

	if envFlags.Executed != nil && !changed("no-status") {
		flags.NoExitCode = !*envFlags.Executed
	}

	return nil
}
