package harness

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// ShellDef describes one of the shells the e2e suite exercises: how to find its
// executable, how to check a generated init script's syntax without running it, and (from
// the smoke/feature test layers, added in later tasks) how to launch it interactively and
// which command reliably fails.
type ShellDef struct {
	// Name is the omp shell name passed to `oh-my-posh init <Name>`.
	Name string

	// Binary is the executable name looked up via LookupShellBinary. A missing binary
	// means the test for this shell should be skipped.
	Binary string

	// SyntaxCheck returns a *exec.Cmd that parses scriptPath without executing it. A
	// non-zero exit code means a syntax error was found.
	SyntaxCheck func(scriptPath string) *exec.Cmd

	// Launch is populated by the smoke/feature test layers (tasks 2-3): it returns the
	// command, arguments and environment needed to boot the shell interactively with
	// scriptPath sourced.
	Launch func(t *testing.T, scriptPath, workDir string) (bin string, args []string, env []string)

	// Fail is populated by the feature test layer (task 3): a command line that, typed
	// interactively, reliably exits non-zero, paired with the exit code oh-my-posh should
	// report in the following prompt.
	Fail ExitCommand
}

// ExitCommand is a command line that reliably exits non-zero when typed interactively,
// together with the exit code the prompt is expected to report afterward (as
// "E2E:<Code>>").
type ExitCommand struct {
	Command string
	Code    int
}

// platformExitCommand picks windows or unix depending on the current OS, so a ShellDef
// can declare one Fail command that differs between Windows and Unix-like hosts.
func platformExitCommand(windows, unix ExitCommand) ExitCommand {
	if runtime.GOOS == "windows" {
		return windows
	}

	return unix
}

// Shells is the per-shell definition table for the big five shells the e2e suite targets.
var Shells = []ShellDef{
	{
		Name:   "bash",
		Binary: "bash",
		SyntaxCheck: func(scriptPath string) *exec.Cmd {
			bin, err := LookupShellBinary("bash")
			if err != nil {
				bin = "bash"
			}

			return exec.Command(bin, "-n", scriptPath)
		},
		Launch: func(t *testing.T, scriptPath, workDir string) (string, []string, []string) {
			rcPath := filepath.Join(workDir, "rc")
			writeLaunchFile(t, rcPath, fmt.Sprintf("source '%s'\n", scriptPath))

			return "bash", []string{"--noprofile", "--rcfile", rcPath, "-i"}, os.Environ()
		},
		Fail: ExitCommand{Command: "false", Code: 1},
	},
	{
		Name:   "zsh",
		Binary: "zsh",
		SyntaxCheck: func(scriptPath string) *exec.Cmd {
			return exec.Command("zsh", "-n", scriptPath)
		},
		Launch: func(t *testing.T, scriptPath, workDir string) (string, []string, []string) {
			writeLaunchFile(t, filepath.Join(workDir, ".zshrc"), fmt.Sprintf("source '%s'\n", scriptPath))

			env := append(os.Environ(), "ZDOTDIR="+workDir)

			return "zsh", []string{"-d", "-i"}, env
		},
		Fail: ExitCommand{Command: "false", Code: 1},
	},
	{
		Name:   "fish",
		Binary: "fish",
		SyntaxCheck: func(scriptPath string) *exec.Cmd {
			return exec.Command("fish", "--no-execute", scriptPath)
		},
		Launch: func(t *testing.T, scriptPath, workDir string) (string, []string, []string) {
			configPath := filepath.Join(workDir, "fish", "config.fish")
			writeLaunchFile(t, configPath, fmt.Sprintf("source '%s'\n", scriptPath))

			env := append(os.Environ(), "XDG_CONFIG_HOME="+workDir)

			return "fish", []string{"-i"}, env
		},
		Fail: ExitCommand{Command: "false", Code: 1},
	},
	{
		Name:   "pwsh",
		Binary: "pwsh",
		SyntaxCheck: func(scriptPath string) *exec.Cmd {
			script := `$errs = $null; [System.Management.Automation.Language.Parser]::ParseFile('` +
				scriptPath + `', [ref]$null, [ref]$errs) | Out-Null; exit $errs.Count`
			return exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", script)
		},
		Launch: func(t *testing.T, scriptPath, workDir string) (string, []string, []string) {
			return "pwsh", []string{"-NoLogo", "-NoProfile", "-NoExit", "-Command", fmt.Sprintf(". '%s'", scriptPath)}, os.Environ()
		},
		Fail: platformExitCommand(
			ExitCommand{Command: `cmd /c "exit 42"`, Code: 42},
			ExitCommand{Command: "/bin/false", Code: 1},
		),
	},
	{
		Name:   "nu",
		Binary: "nu",
		SyntaxCheck: func(scriptPath string) *exec.Cmd {
			script := `exit (if (nu-check '` + scriptPath + `') { 0 } else { 1 })`
			return exec.Command("nu", "--no-config-file", "--commands", script)
		},
		Launch: func(t *testing.T, scriptPath, workDir string) (string, []string, []string) {
			// nu's `source` requires forward slashes even on Windows.
			sourcePath := filepath.ToSlash(scriptPath)

			configPath := filepath.Join(workDir, "config.nu")
			writeLaunchFile(t, configPath, fmt.Sprintf("source '%s'\n", sourcePath))

			envConfigPath := filepath.Join(workDir, "env.nu")
			writeLaunchFile(t, envConfigPath, "")

			// nu autoloads every *.nu file under $nu.vendor-autoload-dirs unconditionally,
			// regardless of --config/--env-config, and runs it after config.nu. On a
			// machine that already has oh-my-posh's nu integration installed as a vendor
			// autoload script (e.g. via its own `oh-my-posh init nu`), that script would
			// otherwise load after ours and silently clobber our test PROMPT_COMMAND with
			// the developer's real prompt. Redirecting XDG_DATA_HOME to an empty temp
			// directory points the vendor-autoload lookup at a directory with nothing in
			// it, isolating the session from whatever is installed on the host.
			env := append(os.Environ(), "XDG_DATA_HOME="+filepath.Join(workDir, "xdg-data"))

			return "nu", []string{"-i", "--config", configPath, "--env-config", envConfigPath}, env
		},
		Fail: platformExitCommand(
			ExitCommand{Command: `^cmd /c "exit 42"`, Code: 42},
			ExitCommand{Command: "^false", Code: 1},
		),
	},
}

// writeLaunchFile writes content to path, creating any missing parent directories, and
// fails the test on error.
func writeLaunchFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating directory for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// SupportedOnHost reports whether sh can be launched interactively on the current
// platform. bash, zsh and fish depend on POSIX rc-file semantics and process/job-control
// behavior that Windows does not provide faithfully (msys bash and WSL shells under
// ConPTY are not representative of a real Linux/macOS session), so they are only
// exercised there. pwsh and nu run natively everywhere the e2e suite targets.
func (sh ShellDef) SupportedOnHost() bool {
	switch sh.Name {
	case "bash", "zsh", "fish":
		return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
	default:
		return true
	}
}

// errBashNotFound is returned by LookupShellBinary when "bash" is requested on Windows
// and no usable Git for Windows installation can be found.
var errBashNotFound = errors.New("no usable bash executable found")

// LookupShellBinary resolves the absolute path to a shell's executable.
//
// For every shell except bash on Windows this is exactly exec.LookPath(name). Bash on
// Windows is special-cased: Windows also ships a WSL launcher named bash.exe under
// System32, and since PATH commonly lists System32 before Git for Windows' cmd directory,
// a plain exec.LookPath("bash") resolves to the WSL launcher. That launcher only
// understands Linux-style paths, but every script path this harness hands to a syntax
// checker is a Windows path (e.g. a t.TempDir() path) — invoking it fails with "No such
// file or directory" regardless of whether the script is valid. On Windows, "bash"
// therefore prefers Git for Windows' own bash.exe.
func LookupShellBinary(name string) (string, error) {
	if name == "bash" && runtime.GOOS == "windows" {
		return lookupGitBash()
	}

	return exec.LookPath(name)
}

func lookupGitBash() (string, error) {
	if gitPath, err := exec.LookPath("git"); err == nil {
		// git.exe lives in <gitRoot>\cmd; Git's own bash.exe ships at <gitRoot>\bin and is
		// duplicated at <gitRoot>\usr\bin.
		gitRoot := filepath.Dir(filepath.Dir(gitPath))
		for _, rel := range []string{filepath.Join("bin", "bash.exe"), filepath.Join("usr", "bin", "bash.exe")} {
			candidate := filepath.Join(gitRoot, rel)
			if isFile(candidate) {
				return candidate, nil
			}
		}
	}

	for _, candidate := range []string{
		`C:\Program Files\Git\bin\bash.exe`,
		`C:\Program Files\Git\usr\bin\bash.exe`,
	} {
		if isFile(candidate) {
			return candidate, nil
		}
	}

	return "", errBashNotFound
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
