package shell

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

//go:embed scripts/omp.nu
var nuInit string

// nuStreamingProbe checks for the experimental `commandline set-prompt`
// (nushell/nushell#18660, not yet released).
const nuStreamingProbe = `scope commands | any {|c| $c.name == "commandline set-prompt"} | into string`

// nuStreamingExecutableEnv lets a user pin the exact nu binary to probe,
// bypassing PATH lookup. Useful when multiple nu installs are on PATH (e.g.
// a stock release ahead of a patched build) and the one that happens to
// resolve first isn't the one that will actually run the generated script.
const nuStreamingExecutableEnv = "POSH_NU_STREAMING_EXECUTABLE"

// nuSupportsStreaming shells out to nu to run nuStreamingProbe. Nu parses and
// arity-checks an entire sourced script up front, even inside unreached
// branches, so this has to be decided before any streaming code is emitted
// into the init script - a runtime check inside the script itself cannot
// prevent the parse failure on a nu build without the PR.
func nuSupportsStreaming(env runtime.Environment) bool {
	executable := "nu"
	if custom := env.Getenv(nuStreamingExecutableEnv); len(custom) != 0 {
		executable = custom
	}

	output, err := env.RunCommand(executable, "--commands", nuStreamingProbe)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == "true"
}

func (f Features) Nu() Code {
	switch f {
	case Transient:
		return `$env.TRANSIENT_PROMPT_COMMAND = {|| _omp_get_prompt transient }`
	case Upgrade:
		return "^$_omp_executable upgrade --auto"
	case Notice:
		return "^$_omp_executable notice"
	case Streaming:
		// commandline set-prompt is not part of released Nushell yet (see
		// https://github.com/nushell/nushell/pull/18660). Nushell parses and
		// arity-checks the whole sourced script up front, even inside branches
		// that never run, so this block can only be emitted when streaming is
		// actually enabled - embedding it unconditionally in omp.nu would break
		// init for every Nu user on a release build.
		return `def --wrapped _omp_stream_primary [
    ...args: string
] {
    let execution_time = match $env.CMD_DURATION_MS {
        '0823' => -1
        $ms => { $ms | into int }
    }

    let no_status = if $nu.history-enabled {
        not ($env.POSH_EXECUTED? | default false)
    } else {
        $execution_time < 0
    }

    let stream_args = [
        --shell=nu
        $"--shell-version=($env.POSH_SHELL_VERSION)"
        $"--status=($env.LAST_EXIT_CODE)"
        $"--no-status=($no_status)"
        $"--execution-time=($execution_time)"
        $"--terminal-width=((term size).columns)"
        $"--job-count=(job list | length)"
    ]

    let exe = $_omp_executable

    # Runs in the background so the line editor is never blocked on slow
    # segments: the first resolved record is sent back to the main thread
    # (job 0) via the mailbox below, and every later record - as pending
    # segments resolve - is pushed live via commandline set-prompt.
    let jid = (job spawn {
        mut sent_first = false
        let self_id = (job id)

        for record in (^$exe stream --save-cache ...$stream_args ...$args | bytes split 0x[00]) {
            if ($record | is-empty) {
                continue
            }

            # \x1e-prefixed records carry the transient prompt (see TransientMarker
            # in src/prompt/streaming.go); Nu already renders that synchronously via
            # its own Transient feature, so skip it here.
            if ($record | bytes starts-with 0x[1e]) {
                continue
            }

            let text = ($record | decode utf8)

            if $sent_first {
                commandline set-prompt $text
            } else {
                $text | job send 0 --tag $self_id
                $sent_first = true
            }
        }
    })

    try {
        job recv --tag $jid --timeout 3sec
    } catch {
        # Nothing streamed back in time (binary missing/crashed, or stuck past
        # the configured "streaming" timeout) - fall back to the blocking
        # render so the prompt still shows something instead of hanging.
        _omp_get_prompt primary ...$args
    }
}

$env.PROMPT_COMMAND = {||
    let hist = if $nu.history-enabled { history } else { [] }
    let hist_len = ($hist | length)

    let clear = $nu.history-enabled and (
        ($hist | is-empty)
        or ($hist | last | get command?) == "clear"
    )

    if ($env.SET_POSHCONTEXT? | is-not-empty) {
        do --env $env.SET_POSHCONTEXT
    }

    $env.POSH_EXECUTED = ($nu.history-enabled and ($hist_len > ($env.POSH_LAST_HISTORY_LEN? | default 0)))
    $env.POSH_LAST_HISTORY_LEN = $hist_len

    _omp_stream_primary $"--cleared=($clear)"
}`
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, FTCSMarks, CursorPositioning, Async, KeyHandlers, VIMode:
		fallthrough
	default:
		return ""
	}
}

func quoteNuStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf(`"%s"`, strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(str))
}
