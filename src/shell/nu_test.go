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
$env._omp_serve_url = null
$env._omp_serve_token = null
$env._omp_serve_port_file = null
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
    if ($env._omp_serve_url? | is-not-empty) {
        try { http post --headers [x-omp-token $env._omp_serve_token] --max-time 500ms --content-type application/json $env._omp_serve_url '{"command":"quit"}' | ignore }
    }

    if ($env._omp_serve_in_fifo? | is-not-empty) {
        try { '{"command":"quit"}' + "\n" | save --append --force $env._omp_serve_in_fifo }
    }

    if ($env._omp_serve_reader_jid? | is-not-empty) {
        try { job kill $env._omp_serve_reader_jid }
    }

    if ($env._omp_serve_in_fifo? | is-not-empty) {
        try { rm -f $env._omp_serve_in_fifo $env._omp_serve_out_fifo }
    }

    if ($env._omp_serve_port_file? | is-not-empty) {
        try { rm -f $env._omp_serve_port_file }
    }

    $env._omp_serve_in_fifo = null
    $env._omp_serve_out_fifo = null
    $env._omp_serve_url = null
    $env._omp_serve_token = null
    $env._omp_serve_port_file = null
    $env._omp_serve_reader_jid = null
}

# Starts a persistent "oh-my-posh serve" daemon, so streamed renders no
# longer pay a fresh process-spawn cost per prompt. Two transports: a fifo
# pair on Unix, loopback HTTP on Windows (no mkfifo there, and nu's write
# primitives all fail against Windows named pipes). _omp_serve_render falls
# back to the per-prompt stream on any failure here.
def --env _omp_serve_start [] {
    _omp_serve_stop

    if ($nu.os-info.name == "windows") {
        _omp_serve_start_http
    } else {
        _omp_serve_start_fifo
    }
}

def --env _omp_serve_start_fifo [] {
    if (which mkfifo | is-empty) {
        return false
    }

    let base = ($nu.temp-dir | path join $"omp-serve-(random uuid)")
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

# Windows transport: requests go out as one-shot "http post" calls and
# records come back over a single long-lived "http get /stream" response
# the reader job holds open for the daemon's lifetime - the moral
# equivalent of the fifo pair. The daemon exits when that stream connection
# breaks, which also covers a hard-killed shell.
def --env _omp_serve_start_http [] {
    let port_file = ($nu.temp-dir | path join $"omp-serve-(random uuid).port")
    let nu_exe = $nu.current-exe
    let exe = $_omp_executable

    # Same nested-nu wrapper as the fifo transport: redirects only resolve
    # correctly at a nu process's own top level, never inside a job-spawn
    # closure. NUL is Windows' null device.
    job spawn {
        ^$nu_exe --no-config-file -c $'^($exe) serve --shell=nu --port-file=($port_file) o+e> NUL'
    }

    # The daemon publishes "<port> <token>" once its listener is up
    # (typically well under 100ms). Written atomically (temp + rename), so
    # a non-empty read is always complete.
    mut endpoint = []
    for _ in 1..100 {
        if ($port_file | path exists) {
            let parts = (try { open --raw $port_file | decode utf8 | str trim | split row " " } catch { [] })
            if ($parts | length) >= 2 {
                $endpoint = $parts
                break
            }
        }

        sleep 20ms
    }

    if ($endpoint | is-empty) {
        return false
    }

    let url = $"http://127.0.0.1:($endpoint.0)"
    let token = $endpoint.1

    # Same reader-job contract as the fifo transport (see the reader in
    # _omp_serve_start_fifo), with the /stream response as the record
    # source. The loop body is duplicated on purpose: routing records
    # through a shared def's $in boundary breaks incremental streaming -
    # records only surface after the response completes.
    let reader_jid = (job spawn {
        let self_id = (job id)
        mut cycle_started = -1

        try {
            for rec in (http get --headers [x-omp-token $token] $"($url)/stream" | bytes split 0x[00]) {
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
        } catch { }
    })

    $env._omp_serve_url = $url
    $env._omp_serve_token = $token
    $env._omp_serve_port_file = $port_file
    $env._omp_serve_reader_jid = $reader_jid
    true
}

# Renders the primary prompt through the daemon, (re)starting it on demand.
# Returns null on any failure, in which case the caller falls back to the
# per-prompt stream.
def --env _omp_serve_render [
    clear: bool
] {
    let started = ($env._omp_serve_in_fifo? | is-not-empty) or ($env._omp_serve_url? | is-not-empty)
    if not $started and not (_omp_serve_start) {
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
    mut env_rec = {PATH: ($env.PATH | str join (char esep))}
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

    let write_ok = if ($env._omp_serve_url? | is-not-empty) {
        (try {
            http post --headers [x-omp-token $env._omp_serve_token] --max-time 3sec --content-type application/json $env._omp_serve_url ($req | to json --raw) | ignore
            true
        } catch { false })
    } else {
        (try {
            ($req | to json --raw) + "\n" | save --append $env._omp_serve_in_fifo
            true
        } catch { false })
    }

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

def _omp_spawn_right [] {
    let execution_time = match $env.CMD_DURATION_MS {
        '0823' => -1
        $ms => { $ms | into int }
    }

    let no_status = if $nu.history-enabled {
        not ($env.POSH_EXECUTED? | default false)
    } else {
        $execution_time < 0
    }

    let exe = $_omp_executable
    let args = [
        --shell=nu
        $"--shell-version=($env.POSH_SHELL_VERSION)"
        $"--status=($env.LAST_EXIT_CODE)"
        $"--no-status=($no_status)"
        $"--execution-time=($execution_time)"
        $"--terminal-width=((term size).columns)"
        $"--job-count=(job list | length)"
    ]

    job spawn {
        let self_id = (job id)
        (^$exe print right --save-cache ...$args) | job send 0 --tag $self_id
    }
}

def --env _omp_collect_right [jid: any] {
    $env._omp_right_prompt = (try {
        job recv --tag $jid --timeout 3sec
    } catch {
        $env._omp_right_prompt? | default ''
    })
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

    # The right prompt renders concurrently with the streamed primary: Nu
    # evaluates PROMPT_COMMAND before PROMPT_COMMAND_RIGHT on every prompt
    # update, so by the time the primary's first record is back the right
    # render (a full process spawn, ~35ms on Windows) has already completed
    # in the background instead of adding a second sequential spawn.
    let right_jid = (_omp_spawn_right)
    let prompt = (_omp_stream_primary $clear)
    _omp_collect_right $right_jid
    $prompt
}

$env.PROMPT_COMMAND_RIGHT = {|| $env._omp_right_prompt? | default '' }`

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
