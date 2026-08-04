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
$env._omp_serve_in_fifo = null
$env._omp_serve_out_fifo = null
$env._omp_serve_reader_jid = null
$env._omp_serve_cycle = 0
$env._omp_serve_failures = 0

# Tears down the serve daemon: sends a graceful quit so the real
# "oh-my-posh serve" process (and the nested nu wrapper that launched it -
# see _omp_serve_start) exit on their own, then kills the reader job and
# removes the fifo pair. A plain "job kill" on the daemon job only reaches
# its direct child (the nested nu wrapper), never the real serve process it
# launched, which is why shutdown goes through the protocol instead.
def --env _omp_serve_stop [] {
    if ($env._omp_serve_in_fifo? | is-not-empty) {
        try { '{"command":"quit"}' + "\n" | save --append --force $env._omp_serve_in_fifo }
    }

    if ($env._omp_serve_reader_jid? | is-not-empty) {
        try { job kill $env._omp_serve_reader_jid }
    }

    if ($env._omp_serve_in_fifo? | is-not-empty) {
        try { rm -f $env._omp_serve_in_fifo $env._omp_serve_out_fifo }
    }

    $env._omp_serve_in_fifo = null
    $env._omp_serve_out_fifo = null
    $env._omp_serve_reader_jid = null
}

# Starts a persistent "oh-my-posh serve" daemon over a pair of named pipes,
# so streamed renders no longer pay a fresh process-spawn cost per prompt.
# Unix only (no mkfifo on Windows) - _omp_serve_render falls back to the
# per-prompt stream on any failure here.
def --env _omp_serve_start [] {
    _omp_serve_stop

    if (which mkfifo | is-empty) {
        return false
    }

    let base = ($nu.temp-path | path join $"omp-serve-(random uuid)")
    let in_fifo = $"($base).in"
    let out_fifo = $"($base).out"

    if (^mkfifo $in_fifo | complete | get exit_code) != 0 {
        return false
    }

    if (^mkfifo $out_fifo | complete | get exit_code) != 0 {
        rm -f $in_fifo
        return false
    }

    # The reader job is the sole consumer of the daemon's output fifo for the
    # daemon's whole lifetime (unlike _omp_stream_primary, which spawns one
    # reader per prompt). The FIRST record of a NEW cycle id unblocks
    # _omp_serve_render via job 0; every LATER record of that same cycle
    # (streamed segment updates, the transient refresh) is pushed live via
    # commandline set-prompt - the same sent_first split _omp_stream_primary
    # does, just scoped across the whole daemon lifetime instead of one
    # subprocess per prompt.
    let reader_jid = (job spawn {
        let self_id = (job id)
        mut cycle_started = -1

        for rec in (open --raw $out_fifo | bytes split 0x[00]) {
            if ($rec | is-empty) {
                continue
            }

            let text = ($rec | decode utf8)
            let sep = ($text | str index-of (char --integer 0x1f))
            if $sep < 0 {
                continue
            }

            let id = ($text | str substring 0..<$sep | into int)
            let payload = ($text | str substring ($sep + 1)..)

            # \x1e-prefixed records carry the transient prompt (see
            # TransientMarker in src/prompt/streaming.go); Nu already renders
            # that synchronously via its own Transient feature.
            if ($payload | str starts-with (char --integer 0x1e)) {
                continue
            }

            if $id != $cycle_started {
                $cycle_started = $id
                {id: $id, payload: $payload} | job send 0 --tag $self_id
            } else {
                commandline set-prompt $payload
            }
        }
    })

    let nu_exe = $nu.current-exe
    let exe = $_omp_executable

    # A direct "job spawn { ^exe serve ... out> $fifo }" never actually forks
    # the child process (verified empirically): the out> redirect only
    # resolves correctly when parsed at a nu process's own top level, not
    # inside a job-spawn closure. Routing through a nested nu subprocess
    # sidesteps that.
    job spawn {
        ^$nu_exe --no-config-file -c $'^($exe) serve --shell=nu --request-pipe=($in_fifo) out> ($out_fifo)'
    }

    $env._omp_serve_in_fifo = $in_fifo
    $env._omp_serve_out_fifo = $out_fifo
    $env._omp_serve_reader_jid = $reader_jid
    true
}

# Renders the primary prompt through the daemon, (re)starting it on demand.
# Returns null on any failure, in which case the caller falls back to the
# per-prompt stream.
def --env _omp_serve_render [
    clear: bool
] {
    if ($env._omp_serve_in_fifo? | is-empty) and not (_omp_serve_start) {
        return null
    }

    $env._omp_serve_cycle = ($env._omp_serve_cycle? | default 0) + 1
    let id = $env._omp_serve_cycle

    let execution_time = match $env.CMD_DURATION_MS {
        '0823' => -1
        $ms => { $ms | into int }
    }

    let no_status = if $nu.history-enabled {
        not ($env.POSH_EXECUTED? | default false)
    } else {
        $execution_time < 0
    }

    # $env.PATH is a list in Nu, not a string - it has to be joined before it
    # can round-trip through the map[string]string the daemon decodes "env"
    # into; sending it as-is silently drops the whole request (encoding/json
    # fails to unmarshal a JSON array into a string value, and serve.go
    # ignores malformed lines for forward/backward compatibility).
    mut env_rec = {PATH: ($env.PATH | str join (char --integer 0x3a))}
    for name in ($env | columns | where {|c| $c starts-with "POSH_" }) {
        let value = ($env | get $name)
        if ($value | describe) == "string" {
            $env_rec = ($env_rec | insert $name $value)
        }
    }

    let req = {
        command: "render"
        id: $id
        shell: "nu"
        "shell-version": $env.POSH_SHELL_VERSION
        status: $env.LAST_EXIT_CODE
        "no-status": $no_status
        "execution-time": $execution_time
        "terminal-width": ((term size).columns)
        "job-count": (job list | length)
        pwd: $env.PWD
        cleared: $clear
        wait: false
        env: $env_rec
    }

    let write_ok = (try {
        ($req | to json --raw) + "\n" | save --append $env._omp_serve_in_fifo
        true
    } catch { false })

    if not $write_ok {
        _omp_serve_stop
        return null
    }

    let msg = (try {
        job recv --tag $env._omp_serve_reader_jid --timeout 3sec
    } catch { null })

    if ($msg == null) or ($msg.id != $id) {
        return null
    }

    $msg.payload
}

def --env _omp_stream_primary [
    clear: bool
] {
    # serve daemon: persistent process, no spawn per prompt. After three
    # failures the daemon is left alone for the session and every prompt
    # takes the per-prompt stream below instead.
    if $env._omp_serve_failures < 3 {
        let rendered = (_omp_serve_render $clear)
        if ($rendered | is-not-empty) {
            return $rendered
        }

        $env._omp_serve_failures = $env._omp_serve_failures + 1
    }

    let args = [$"--cleared=($clear)"]

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

    _omp_stream_primary $clear
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
