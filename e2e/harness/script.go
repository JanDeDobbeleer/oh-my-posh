package harness

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// scriptExtensions maps an omp shell name to the file extension its init script should
// be written with, so shell parsers that dispatch on extension behave correctly.
var scriptExtensions = map[string]string{
	"bash": ".sh",
	"zsh":  ".zsh",
	"fish": ".fish",
	"pwsh": ".ps1",
	"nu":   ".nu",
}

// InitScript runs `<omp> init <shellName> --config <cfgPath> --print` with OMP_CACHE_DIR
// pointed at a fresh t.TempDir(), and returns the printed init script text.
func InitScript(t *testing.T, shellName, cfgPath string) string {
	t.Helper()

	bin := Binary(t)

	cmd := exec.Command(bin, "init", shellName, "--config", cfgPath, "--print")
	cmd.Env = append(os.Environ(), "OMP_CACHE_DIR="+t.TempDir())

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("omp init %s --config %s --print failed: %v\n%s", shellName, cfgPath, err, output)
	}

	return string(output)
}

// WriteScript writes script to a temp file named with the extension appropriate for
// shellName (e.g. ".sh" for bash, ".ps1" for pwsh) and returns the absolute path.
func WriteScript(t *testing.T, shellName, script string) string {
	t.Helper()

	ext, ok := scriptExtensions[shellName]
	if !ok {
		t.Fatalf("no script extension known for shell %q", shellName)
	}

	path := filepath.Join(t.TempDir(), "init"+ext)
	if err := os.WriteFile(path, []byte(script), 0o644); err != nil {
		t.Fatalf("writing init script: %v", err)
	}

	return path
}
