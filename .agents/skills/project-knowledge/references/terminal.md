# Terminal

## Windows Terminal and ConPTY

- The "[process exited with code 0] / Ctrl+D to close" text flashing at shell exit is **Windows
  Terminal's own teardown message** (`ConptyConnection::_indicateExitWithStatus`) - visibility is
  a render/close race, never omp output.
- Panes launched via `wt.exe <commandline>` never auto-close under `closeOnExit=automatic`;
  profile-launched tabs do.
- On Windows 11 build 26300 a headless conhost reports cursor 0,0 / blank buffer and drops pty
  pipe input - visual row placement cannot be asserted headlessly, only state-machine behavior.
- pywinpty's ConPTY teardown EOF lags process death by a constant ~5s (harness artifact - it shows
  up in a plain-pwsh control too).
- SendKeys into Windows Terminal is unreliable while the user works - keystrokes get silently
  lost, so a "hung" test may simply never have received its `exit`.

## Encoding

- Under a stock CP437/ACP1252 console, pwsh's `&` call operator mangles UTF-8 program output
  (U+E0B6 becomes U+03B5). Reading a native program's UTF-8 output reliably requires
  `[Console]::OutputEncoding = UTF8` or the Process API.
- PowerShell `| Set-Content -NoNewline` joins piped lines into ONE line and corrupts scripts -
  write files with Git Bash redirection when byte fidelity matters.

## Process facts

- Windows process creation floor is ~70ms; a timer tick is 15.6ms (quantizes every short sleep).
- On Windows there is no SIGPIPE; child lifecycle management must rely on fd closure / stdin EOF.
- On native Linux, process spawns cost 11-16ms - daemon architectures that pay off on Windows can
  be a wash there (see the bash serve revert in [bash](bash.md)).
