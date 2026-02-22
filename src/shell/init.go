package shell

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

const (
	noExe = "echo \"Unable to find Oh My Posh executable\""
)

var (
	// identify ble.sh by validating the existence of BLE_SESSION_ID
	bashBLEsession bool
)

func getExecutablePath(env runtime.Environment) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	_, msix := cache.PackageFamilyName()
	if msix || env.Flags().Strict {
		return path.Base(executable), nil
	}

	// On Windows, it fails when the executable is called in MSYS2 for example
	// which uses unix style paths to resolve the executable's location.
	// PowerShell knows how to resolve both, so we can swap this without any issue.
	if env.GOOS() == runtime.WINDOWS {
		executable = strings.ReplaceAll(executable, "\\", "/")
	}

	return executable, nil
}

// Init returns the command to initialize oh-my-posh for the shell.
// It writes the init script to the appropriate location and returns
// a source command or wrapper command depending on the shell.
// For Nu shell, it writes to the autoload directory and returns empty.
// For PWSH/Elvish, it returns a wrapper command that re-invokes oh-my-posh.
func Init(env runtime.Environment, feats Features) string {
	switch env.Flags().Shell {
	case PWSH:
		if !env.Flags().Eval {
			return generateAndSourceScript(env, feats)
		}

		return recurseInitCommand(env)
	case ELVISH:
		return recurseInitCommand(env)
	case NU:
		return initNu(env, feats)
	case ZSH, BASH, FISH, CMD, XONSH:
		return generateAndSourceScript(env, feats)
	default:
		return fmt.Sprintf(`echo "%s is not supported by Oh My Posh"`, env.Flags().Shell)
	}
}

// Script returns the init script content directly.
// This is used by the --print flag to output the script to stdout.
func Script(env runtime.Environment, feats Features) string {
	script := generateScript(env, feats)
	return fmt.Sprintf("%s\n%s", sessionScript(env.Flags().Shell), script)
}

// Debug writes the init script and returns debug information.
// This is used by the --debug flag.
func Debug(env runtime.Environment, feats Features, startTime *time.Time) string {
	script := generateScript(env, feats)

	log.Debug(script)

	if _, err := writeScript(env, script); err != nil {
		log.Error(err)
	}

	return printDebugInfo(env, startTime)
}

// recurseInitCommand returns a wrapper command that re-invokes oh-my-posh
// with --print. This is used by PWSH and Elvish which eval the script.
func recurseInitCommand(env runtime.Environment) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	var additionalParams string

	if env.Flags().Strict {
		additionalParams += " --strict"
	}

	if env.Flags().Eval {
		additionalParams += " --eval"
	}

	config := quotePwshOrElvishStr(env.Flags().ConfigPath)
	executable = quotePwshOrElvishStr(executable)

	var command string

	switch env.Flags().Shell {
	case PWSH:
		command = "(@(& %s init %s --config=%s --print%s) -join \"`n\") | Invoke-Expression"
	case ELVISH:
		command = "eval ((external %s) init %s --config=%s --print%s | slurp)"
	}

	return fmt.Sprintf(command, executable, env.Flags().Shell, config, additionalParams)
}

// generateAndSourceScript writes the init script to the cache and returns a source command.
func generateAndSourceScript(env runtime.Environment, feats Features) string {
	async := feats&Async != 0

	if scriptPath, ok := hasScript(env); ok {
		return sourceCommand(env, scriptPath, async)
	}

	script := generateScript(env, feats)

	log.Debug(script)

	scriptPath, err := writeScript(env, script)
	if err != nil {
		return fmt.Sprintf("echo \"Failed to write init script: %s\"", err.Error())
	}

	return sourceCommand(env, scriptPath, async)
}

// initNu writes the init script to Nu's autoload directory.
// It returns empty since Nu automatically loads from the autoload directory.
func initNu(env runtime.Environment, feats Features) string {
	script := generateNuScript(env, feats)

	scriptPath, err := writeScript(env, script)
	if err != nil {
		return fmt.Sprintf("echo \"Failed to write init script: %s\"", err.Error())
	}

	log.Debug("nu init script written to:", scriptPath)

	return ""
}

// generateScript generates the init script content for the current shell.
func generateScript(env runtime.Environment, feats Features) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	bashBLEsession = len(env.Getenv("BLE_SESSION_ID")) != 0

	var script string

	switch env.Flags().Shell {
	case PWSH:
		executable = quotePwshOrElvishStr(executable)
		script = pwshInit
	case ZSH:
		executable = QuotePosixStr(executable)
		script = zshInit
	case BASH:
		executable = QuotePosixStr(executable)
		script = bashInit
	case FISH:
		executable = quoteFishStr(executable)
		script = fishInit
	case CMD:
		executable = escapeLuaStr(executable)
		script = cmdInit
	case NU:
		executable = quoteNuStr(executable)
		script = nuInit
	case ELVISH:
		executable = quotePwshOrElvishStr(executable)
		script = elvishInit
	case XONSH:
		executable = quotePythonStr(executable)
		script = xonshInit
	default:
		return fmt.Sprintf("echo \"No initialization script available for %s\"", env.Flags().Shell)
	}

	// Remove UTF-8 BOM if present, as it can cause issues in some shells.
	script = strings.TrimPrefix(script, "\xef\xbb\xbf")

	init := strings.NewReplacer(
		"::OMP::", executable,
		"::SESSION_ID::", cache.SessionID(),
	).Replace(script)

	return feats.Lines(env.Flags().Shell).String(init)
}

// generateNuScript generates the init script content specifically for Nu shell.
func generateNuScript(env runtime.Environment, feats Features) string {
	executable, err := getExecutablePath(env)
	if err != nil {
		return noExe
	}

	executable = quoteNuStr(executable)

	init := strings.NewReplacer(
		"::OMP::", executable,
		"::SESSION_ID::", cache.SessionID(),
	).Replace(nuInit)

	return feats.Lines(NU).String(init)
}

// sourceCommand returns the command to source the init script.
func sourceCommand(env runtime.Environment, scriptPath string, async bool) string {
	if env.IsCygwin() {
		var err error
		scriptPath, err = env.RunCommand("cygpath", "-u", scriptPath)
		if err != nil {
			log.Error(err)
			return fmt.Sprintf("echo \"Failed to convert Cygwin path due to %s\"", err.Error())
		}
	}

	script := sessionScript(env.Flags().Shell)

	if async {
		return script + sourceCommandAsync(env.Flags().Shell, scriptPath)
	}

	switch env.Flags().Shell {
	case PWSH:
		script += fmt.Sprintf("& %s", quotePwshOrElvishStr(scriptPath))
	case ZSH, BASH:
		script += fmt.Sprintf("source %s", QuotePosixStr(scriptPath))
	case XONSH:
		script += fmt.Sprintf("source %s", quotePythonStr(scriptPath))
	case FISH:
		script += fmt.Sprintf("source %s", quoteFishStr(scriptPath))
	case ELVISH:
		script += fmt.Sprintf("eval (slurp < %s)", quotePwshOrElvishStr(scriptPath))
	case CMD:
		script += fmt.Sprintf(`load(io.open('%s', "r"):read("*a"))()`, escapeLuaStr(scriptPath))
	default:
		return fmt.Sprintf("echo \"No source command available for %s\"", env.Flags().Shell)
	}

	return script
}

// sourceCommandAsync returns the async source command for supported shells.
func sourceCommandAsync(shell, scriptPath string) string {
	switch shell {
	case PWSH:
		return fmt.Sprintf("function prompt() { & %s }", quotePwshOrElvishStr(scriptPath))
	case ZSH:
		return fmt.Sprintf("precmd() { source %s }", QuotePosixStr(scriptPath))
	case BASH:
		command := fmt.Sprintf("source %s", QuotePosixStr(scriptPath))
		return fmt.Sprintf("PROMPT_COMMAND=%s", QuotePosixStr(command))
	case FISH:
		return fmt.Sprintf("function fish_prompt; source %s; end", quoteFishStr(scriptPath))
	default:
		return ""
	}
}

func printDebugInfo(env runtime.Environment, startTime *time.Time) string {
	builder := text.NewBuilder()

	builder.WriteString(fmt.Sprintf("\n%s %s\n", log.Text("Init duration:").Green().Bold().Plain(), time.Since(*startTime)))

	builder.WriteString(log.Text("\n\nLogs:\n\n").Green().Bold().Plain().String())
	builder.WriteString(env.Logs())

	return builder.String()
}

func sessionScript(shell string) string {
	switch shell {
	case PWSH:
		return fmt.Sprintf("$env:POSH_SESSION_ID = \"%s\";", cache.SessionID())
	case ZSH, BASH:
		return fmt.Sprintf("export POSH_SESSION_ID=\"%s\";", cache.SessionID())
	case XONSH:
		return fmt.Sprintf("$POSH_SESSION_ID = \"%s\";", cache.SessionID())
	case FISH:
		return fmt.Sprintf("set --export --global POSH_SESSION_ID \"%s\";", cache.SessionID())
	case ELVISH:
		return fmt.Sprintf("set-env POSH_SESSION_ID \"%s\";", cache.SessionID())
	case CMD:
		return fmt.Sprintf(`os.setenv('POSH_SESSION_ID', '%s');`, cache.SessionID())
	}
	return ""
}
