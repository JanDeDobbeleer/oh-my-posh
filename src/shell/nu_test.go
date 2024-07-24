package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNuFeatures(t *testing.T) {
	got := allFeatures.Lines(NU).String("// these are the features")

	//nolint: lll
	want := `// these are the features
$env.TRANSIENT_PROMPT_COMMAND = { ^$_omp_executable print transient $"--config=($env.POSH_THEME)" --shell=nu $"--shell-version=($env.POSH_SHELL_VERSION)" $"--execution-time=(posh_cmd_duration)" $"--status=($env.LAST_EXIT_CODE)" $"--terminal-width=(posh_width)" }
^$_omp_executable upgrade
^$_omp_executable notice`

	assert.Equal(t, want, got)
}
