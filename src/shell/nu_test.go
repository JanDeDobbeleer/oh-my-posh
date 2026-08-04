package shell

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNuFeatures(t *testing.T) {
	got := allFeatures.Lines(NU).String("// these are the features")

	want := `// these are the features
$env.TRANSIENT_PROMPT_COMMAND = {|| _omp_get_prompt transient }
^$_omp_executable upgrade --auto
^$_omp_executable notice
def --wrapped _omp_stream_primary [
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

	assert.Equal(t, want, got)
}

func TestQuoteNuStr(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{str: "", expected: "''"},
		{str: `/tmp/"omp's dir"/oh-my-posh`, expected: `"/tmp/\"omp's dir\"/oh-my-posh"`},
		{str: `C:/tmp\omp's dir/oh-my-posh.exe`, expected: `"C:/tmp\\omp's dir/oh-my-posh.exe"`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, quoteNuStr(tc.str), fmt.Sprintf("quoteNuStr: %s", tc.str))
	}
}
